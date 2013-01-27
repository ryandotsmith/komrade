package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Job struct {
	Id        string                 `json:"id"`
	QueueId   string                 `json:",omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	LockedAt  time.Time              `json:"locked_at"`
	Payload   map[string]interface{} `json:"payload"`
}

func GetJobs(queueId, limit string) ([]*Job, error) {
	pg, err := sql.Open("postgres", pgurl)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return nil, err
	}
	defer pg.Close()

	s := "select id, created_at, locked_at, payload from lock_jobs($1,$2)"
	rows, err := pg.Query(s, queueId, limit)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
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
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (j *Job) Get() bool {
	pg, err := sql.Open("postgres", pgurl)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return false
	}
	defer pg.Close()
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
	pg, err := sql.Open("postgres", pgurl)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}
	defer pg.Close()

	payload, err := json.Marshal(j.Payload)
	if err != nil {
		return err
	}

	txn, err := pg.Begin()
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}

	s := "insert into jobs (queue, id, payload) values ($1,$2,$3)"
	_, err = txn.Exec(s, j.QueueId, j.Id, string(payload))
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}
	fmt.Printf("measure=jobs.insert id=%s\n", j.Id)

	// Count number of jobs passing through the queue.
	s = "update queues set job_count = job_count + 1 "
	s += "where token = $1"
	_, err = txn.Exec(s, j.QueueId)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}

	err = txn.Commit()
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}
	return nil
}

func (j *Job) Delete() error {
	pg, err := sql.Open("postgres", pgurl)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}
	defer pg.Close()
	s := "delete from jobs where id = $1 returning payload"
	rows, err := pg.Query(s, j.Id)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}
	fmt.Printf("measure=jobs.delete id=%s\n", j.Id)
	defer rows.Close()
	rows.Next()
	var tmp []byte
	if err = rows.Scan(&tmp); err != nil {
		return err
	}
	if err = json.Unmarshal(tmp, &j.Payload); err != nil {
		return err
	}
	return nil
}
