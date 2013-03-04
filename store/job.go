package store

import (
	"encoding/json"
	"fmt"
	"time"
)

type Job struct {
	Id        string                 `json:"id"`
	QueueId   string                 `json:",omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	LockedAt  time.Time              `json:"locked_at"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

func DeleteAllJobs(queueId string) (int64, error) {
	s := "delete from jobs where queue = $1"
	res, err := pg.Exec(s, queueId)
	if err != nil {
		return 0, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	fmt.Printf("measure=%q val=%d\n", "jobs.delelte-all", count)
	return count, nil
}

func GetJobs(queueId, limit string) ([]*Job, error) {
	s := "select id, created_at, locked_at, payload from lock_jobs($1,$2)"
	rows, err := pg.Query(s, queueId, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := make([]*Job, 0)
	for rows.Next() {
		j := new(Job)
		var tmp []byte
		err := rows.Scan(&j.Id, &j.CreatedAt, &j.LockedAt, &tmp)
		if err != nil {
			fmt.Printf("at=error error=%s\n", err)
			continue
		}
		json.Unmarshal(tmp, &j.Payload)
		if err != nil {
			fmt.Printf("at=error error=%s\n", err)
			continue
		}
		go Record(dequeueEvent, queueId, "")
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (j *Job) Get() bool {
	s := "select payload from jobs where id = $1"
	rows, err := pg.Query(s, j.Id)
	if err != nil {
		return false
	}
	defer rows.Close()
	rows.Next()
	var tmp []byte
	if err = rows.Scan(&tmp); err != nil {
		return false
	}
	if err = json.Unmarshal(tmp, &j.Payload); err != nil {
		return false
	}
	return true
}

func (j *Job) Insert() error {
	payload, err := json.Marshal(j.Payload)
	if err != nil {
		return err
	}

	tx, err := pg.Begin()
	if err != nil {
		return err
	}

	s := "insert into jobs (queue, id, payload) values ($1,$2,$3)"
	_, err = tx.Exec(s, j.QueueId, j.Id, string(payload))
	if err != nil {
		tx.Rollback()
		return err
	}

	s = "update queues set length = length + 1 where token = $1"
	_, err = tx.Exec(s, j.QueueId)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	fmt.Printf("measure=jobs.insert id=%s\n", j.Id)
	go Record(enqueueEvent, j.QueueId, j.Id)
	return nil
}

func (j *Job) HeartBeat() error {
	defer fmt.Printf("measure=jobs.heartbeat id=%s\n", j.Id)
	s := "update jobs set "
	s += "heartbeat = now(), "
	s += "heartbeat_count = heartbeat_count + 1 "
	s += "where id = $1 and queue = $2"
	_, err := pg.Exec(s, j.Id, j.QueueId)
	if err != nil {
		return err
	}
	return nil
}

func (j *Job) Delete() error {
	tx, err := pg.Begin()
	if err != nil {
		return err
	}

	s := "delete from jobs where id = $1"
	_, err = tx.Exec(s, j.Id)
	if err != nil {
		tx.Rollback()
		return err
	}

	s = "update queues set length = length - 1 where token = $1"
	_, err = tx.Exec(s, j.QueueId)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	fmt.Printf("measure=jobs.delete id=%s\n", j.Id)
	go Record(deleteEvent, j.QueueId, j.Id)
	return nil
}
