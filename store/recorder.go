package store

import (
	"fmt"
)

const (
	enqueueEvent = 0
	dequeueEvent = 1
	deleteEvent  = 2
)

func Record(event int, queueId, jobId string) {
	switch event {
	case enqueueEvent:
		fmt.Printf("at=enqueue\n")
	case dequeueEvent:
		fmt.Printf("at=dequeue\n")
	case deleteEvent:
		fmt.Printf("at=delete\n")
	}
}
