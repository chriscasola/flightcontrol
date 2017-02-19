package flightcontrol

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/chriscasola/testdata"
)

type mockJob struct {
	numDos  int
	numLock sync.Mutex
}

func (t *mockJob) Do() {
	t.numLock.Lock()
	t.numDos++
	t.numLock.Unlock()
}

func newMockJob() *mockJob {
	return &mockJob{numDos: 0}
}

func TestWorker(t *testing.T) {
	waiter := &sync.WaitGroup{}
	workerPool := make(chan *worker)
	worker := newWorker(workerPool, waiter)
	worker.Start()

	job := newMockJob()

	// wait for the worker to start
	availableWorker := <-workerPool

	// give the job to the worker
	waiter.Add(1)
	availableWorker.JobChannel <- job

	// wait for the worker to finish
	availableWorker = <-workerPool

	// give the job to the worker
	waiter.Add(1)
	availableWorker.JobChannel <- job

	// wait for the worker to finish
	availableWorker = <-workerPool

	if job.numDos != 2 {
		t.Error("Expected the job to be done twice")
	}

	worker.Stop()
}

func TestDispatcher(t *testing.T) {
	dispatcher := NewDispatcher(4, 100)

	if cap(dispatcher.workerPool) != 4 {
		t.Error("Expected worker pool to be of size 4")
	}

	if cap(dispatcher.jobQueue) != 100 {
		t.Error("Expected job queue to be of size 100")
	}

	dispatcher.Start()
	job := newMockJob()

	for i := 0; i < 50; i++ {
		dispatcher.QueueJob(job)
	}

	dispatcher.Flush()

	if job.numDos != 50 {
		t.Error("Expected 50 jobs to have been run")
	}

	dispatcher.Stop()

	time.Sleep(200 * time.Millisecond)

	if len(dispatcher.workerPool) != 0 {
		t.Error("Expected workers to have stopped")
	}
}

type mockExpensiveJob struct {
	largeJSON []byte
	result    interface{}
	wg        *sync.WaitGroup
}

func (j *mockExpensiveJob) Do() {
	json.Unmarshal(j.largeJSON, &j.result)
	j.wg.Done()
}

func newExpensiveJob(someJSON []byte, wg *sync.WaitGroup) *mockExpensiveJob {
	return &mockExpensiveJob{largeJSON: someJSON, wg: wg}
}

func BenchmarkDispatcher(b *testing.B) {
	testdata.WriteTestData(100000, "test_data.json")
	testFile, _ := ioutil.ReadFile("test_data.json")

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		var wg sync.WaitGroup
		dispatcher := NewDispatcher(4, 100)
		dispatcher.Start()

		for i := 0; i < 25; i++ {
			someJob := newExpensiveJob(testFile, &wg)
			wg.Add(1)
			dispatcher.QueueJob(someJob)
		}

		wg.Wait()
		dispatcher.Stop()
	}

	b.StopTimer()

	os.Remove("test_data.json")
}

type dataRow struct {
	ID    string
	Cells []float32
}

func BenchmarkJSONParse(b *testing.B) {
	testdata.WriteTestData(1000000, "test_data_large.json")
	testFile, _ := ioutil.ReadFile("test_data_large.json")

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		var result []dataRow
		json.Unmarshal(testFile, &result)
	}

	b.StopTimer()

	os.Remove("test_data_large.json")
}

func ExampleStart() {
	dispatcher := NewDispatcher(4, 100)

	dispatcher.Start()

	// add 50 jobs to the queue
	job := newMockJob()
	for i := 0; i < 50; i++ {
		dispatcher.QueueJob(job)
	}

	// wait for all jobs to finish
	dispatcher.Flush()

	fmt.Println(job.numDos)

	dispatcher.Stop()

	// Output:
	// 50
}
