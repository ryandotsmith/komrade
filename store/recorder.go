package store

import (
	"fmt"
)

const (
	enqueueEvent = 0
	dequeueEvent = 1
	deleteEvent  = 2
	errorEvent   = 3
)

func Record(event int, queueId, jobId string) {
	var err error
	if len(jobId) == 0 {
		s := "insert into metabolism_reports (queue, action) "
		s += "values ($1,$2)"
		_, err = pg.Exec(s, queueId, event)
	} else {
		s := "insert into metabolism_reports (queue, job, action) "
		s += "values ($1,$2,$3)"
		_, err = pg.Exec(s, queueId, jobId, event)
	}
	if err != nil {
		fmt.Printf("measure=%q error=%s event=%d queue=%s\n",
			"record-event-error", err, event, queueId)
	}
}
