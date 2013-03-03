package store

import (
	"encoding/json"
	"fmt"
)

type FailedJob struct {
	Id         string                 `json:"id"`
	Payload    map[string]interface{} `json:"payload"`
	QueueId    string                 `json:"queue_id"`
	JobId      string                 `json:"job_id"`
	JobPayload map[string]interface{} `json:"job_payload"`
}

func (f *FailedJob) Get() bool {
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
	txn, err := pg.Begin()
	if err != nil {
		return err
	}

	payload, err := json.Marshal(f.Payload)
	if err != nil {
		return err
	}

	jobPayload, err := json.Marshal(f.JobPayload)
	if err != nil {
		return err
	}

	s := "insert into failed_jobs (queue, job_id, job_payload, id, payload) "
	s += "values($1,$2,$3, $4)"
	_, err = txn.Exec(s, f.QueueId, f.JobId, string(jobPayload),
		f.Id, string(payload))
	if err != nil {
		return err
	}
	fmt.Printf("measure=failed_jobs.insert id=%s\n", f.Id)

	// This way we can quickly find jobs with high failure count.
	// Also, this job should be available for work again.
	s = "update jobs set failed_count = (failed_count + 1), locked_at = null "
	s += "where jobs.id = $1"
	rows, err := txn.Query(s, f.JobId)
	if err != nil {
		return err
	}
	rows.Close()

	err = txn.Commit()
	if err != nil {
		return err
	}
	go Record(errorEvent, f.QueueId, f.JobId)
	return nil
}
