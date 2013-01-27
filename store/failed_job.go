package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type FailedJob struct {
	Id      string                 `json:"id"`
	JobId   string                 `json:"id"`
	QueueId string                 `json:"id"`
	Payload map[string]interface{} `json:"payload"`
}

func (f *FailedJob) Get() bool {
	pg, err := sql.Open("postgres", pgurl)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return false
	}
	defer pg.Close()

	s := "select payload from failed_jobs where id = $1"
	rows, err := pg.Query(s, f.Id)
	if err != nil {
		return false
	}
	defer rows.Close()
	rows.Next()
	var tmp []byte
	if err = rows.Scan(&tmp); err != nil {
		return false
	}
	if err = json.Unmarshal(tmp, &f.Payload); err != nil {
		return false
	}
	return true
}

func (f *FailedJob) Insert() error {
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

	payload, err := json.Marshal(f.Payload)
	if err != nil {
		return err
	}

	s := "insert into failed_jobs (queue, job_id, id, payload) "
	s += "values($1,$2,$3, $4)"
	_, err = txn.Exec(s, f.QueueId, f.JobId, f.Id, string(payload))
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}
	fmt.Printf("measure=failed_jobs.insert id=%s\n", f.Id)

	// This way we can quickly find jobs with high failure count.
	// Also, this job should be available for work again.
	s = "update jobs set failed_count = (failed_count + 1), locked_at = null "
	s += "where jobs.id = $1"
	rows, err := txn.Query(s, f.JobId)
	if err != nil {
		fmt.Printf("at=error error=%s\n", err)
		return err
	}
	rows.Close()

	rows, err = txn.Query("select update_error_counter($1)", f.QueueId)
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
