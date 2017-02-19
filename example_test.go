package flightcontrol_test

import (
	"fmt"
	"sync/atomic"

	"github.com/chriscasola/flightcontrol"
)

type NumberJob struct {
	num int32
}

func (n *NumberJob) Do() {
	atomic.AddInt32(&n.num, 1)
}

func NewNumberJob() *NumberJob {
	return &NumberJob{num: 0}
}

func Example() {
	dispatcher := flightcontrol.NewDispatcher(4, 100)

	dispatcher.Start()

	// add a job to the queue 50 times
	job := NewNumberJob()
	for i := 0; i < 50; i++ {
		dispatcher.QueueJob(job)
	}

	// wait for all jobs to finish
	dispatcher.Flush()

	fmt.Println(job.num)

	dispatcher.Stop()
	// Output: 50
}
