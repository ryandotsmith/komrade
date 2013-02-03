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

func DeleteAllJobs(queueId string) (int64, error) {
	pg, err := sql.Open("postgres", pgurl)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return 0, err
	}
	defer pg.Close()
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
	rows, err := txn.Query(s, j.QueueId, j.Id, string(payload))
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}
	fmt.Printf("measure=jobs.insert id=%s\n", j.Id)
	rows.Close()

	rows, err = txn.Query("select update_in_counter($1)", j.QueueId)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}
	rows.Close()

	err = txn.Commit()
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}

	return nil
}

func (j *Job) HeartBeat() error {
	defer fmt.Printf("measure=jobs.heartbeat id=%s\n", j.Id)

	pg, err := sql.Open("postgres", pgurl)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}
	defer pg.Close()

	s := "update jobs set "
	s += "heartbeat = now(), "
	s += "heartbeat_count = heartbeat_count + 1 "
	s += "where id = $1 and queue = $2"
	_, err = pg.Exec(s, j.Id, j.QueueId)
	if err != nil {
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

	txn, err := pg.Begin()
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}

	s := "delete from jobs where id = $1 returning payload"
	rows, err := txn.Query(s, j.Id)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}

	fmt.Printf("measure=jobs.delete id=%s\n", j.Id)
	rows.Next()
	var tmp []byte
	if err = rows.Scan(&tmp); err != nil {
		return err
	}
	if err = json.Unmarshal(tmp, &j.Payload); err != nil {
		return err
	}
	rows.Close()

	rows, err = txn.Query("select update_out_counter($1)", j.QueueId)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}
	rows.Close()

	err = txn.Commit()
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}
	return nil
}
