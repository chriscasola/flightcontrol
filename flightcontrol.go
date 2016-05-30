package flightcontrol

// Job is a job to be run
type Job interface {
	Do()
}

// Worker can execute jobs
type Worker struct {
	WorkerPool chan *Worker
	JobChannel chan Job
	quit       chan bool
}

// NewWorker creates a new worker
func NewWorker(workerPool chan *Worker) *Worker {
	return &Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
	}
}

// Start method starts the run loop for a worker
func (w *Worker) Start() {
	go func() {
		for {
			w.WorkerPool <- w

			select {
			case job := <-w.JobChannel:
				job.Do()
			case <-w.quit:
				return
			}
		}
	}()
}

// Stop stops the given worker
func (w *Worker) Stop() {
	w.quit <- true
}

// Dispatcher spawns workers and coordinates
type Dispatcher struct {
	workerPool chan *Worker
	jobQueue   chan Job
	stop       chan bool
}

// NewDispatcher creates a dispatcher with the given number of workers
func NewDispatcher(maxWorkers int, maxJobsInQueue int) *Dispatcher {
	pool := make(chan *Worker, maxWorkers)
	jobQueue := make(chan Job, maxJobsInQueue)
	stop := make(chan bool)
	return &Dispatcher{workerPool: pool, jobQueue: jobQueue, stop: stop}
}

// Run starts the dispatcher which in turn starts all the workers
func (d *Dispatcher) Run() {
	for i := 0; i < cap(d.workerPool); i++ {
		worker := NewWorker(d.workerPool)
		worker.Start()
	}

	go d.dispatch()
}

// QueueJob tells the dispatcher to schedule a job
func (d *Dispatcher) QueueJob(j Job) {
	d.jobQueue <- j
}

// Stop stops the dispatch loop
func (d *Dispatcher) Stop() {
	d.stop <- true
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
