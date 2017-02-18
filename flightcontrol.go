package flightcontrol

import "sync"

// Job is a job to be run
type Job interface {
	Do()
}

// Worker can execute jobs
type worker struct {
	WorkerPool chan *worker
	JobChannel chan Job
	quit       chan bool
	waiter     *sync.WaitGroup
}

// NewWorker creates a new worker
func newWorker(workerPool chan *worker, waiter *sync.WaitGroup) *worker {
	return &worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
		waiter:     waiter,
	}
}

// Start method starts the run loop for a worker
func (w *worker) Start() {
	go func() {
		for {
			w.WorkerPool <- w

			select {
			case job := <-w.JobChannel:
				job.Do()
				w.waiter.Done()
			case <-w.quit:
				return
			}
		}
	}()
}

// Stop stops the given worker
func (w *worker) Stop() {
	w.quit <- true
}

// Dispatcher spawns workers and coordinates
type Dispatcher struct {
	workerPool chan *worker
	jobQueue   chan Job
	stop       chan bool
	waiter     *sync.WaitGroup
}

// NewDispatcher creates a dispatcher with the given number of workers
func NewDispatcher(maxWorkers int, maxJobsInQueue int) *Dispatcher {
	pool := make(chan *worker, maxWorkers)
	jobQueue := make(chan Job, maxJobsInQueue)
	stop := make(chan bool)
	return &Dispatcher{workerPool: pool, jobQueue: jobQueue, stop: stop, waiter: &sync.WaitGroup{}}
}

// Run starts the dispatcher which in turn starts all the workers
func (d *Dispatcher) Run() {
	for i := 0; i < cap(d.workerPool); i++ {
		worker := newWorker(d.workerPool, d.waiter)
		worker.Start()
	}

	go d.dispatch()
}

// QueueJob tells the dispatcher to schedule a job
func (d *Dispatcher) QueueJob(j Job) {
	d.waiter.Add(1)
	d.jobQueue <- j
}

// Stop stops the dispatch loop
func (d *Dispatcher) Stop() {
	d.stop <- true
}

// Flush waits for the job queue to become empty
func (d *Dispatcher) Flush() {
	d.waiter.Wait()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			availableWorker := <-d.workerPool
			availableWorker.JobChannel <- job
		case <-d.stop:
			d.stopWorkers()
			return
		}
	}
}

func (d *Dispatcher) stopWorkers() {
	for len(d.workerPool) > 0 {
		worker := <-d.workerPool
		worker.Stop()
	}
}
