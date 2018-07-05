package dispatch

import (
	"log"

	"github.com/alecholmez/workerPool/work"
)

// Dispatcher is the queue manager for the worker pool
type Dispatcher struct {
	WorkerPool chan chan struct{}
	JobQueue   chan struct{}
	maxWorkers int
}

// NewDispatcher will create a new dispatcher with the appropriate amount of workers
func NewDispatcher(maxWorkers int, jobQueue chan struct{}) *Dispatcher {
	pool := make(chan chan struct{}, maxWorkers)
	return &Dispatcher{WorkerPool: pool, maxWorkers: maxWorkers, JobQueue: jobQueue}
}

// Run will start the dispatcher and create our queue
func (d *Dispatcher) Run() {
	for i := 0; i < d.maxWorkers; i++ {
		worker := work.NewWorker(d.WorkerPool)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.JobQueue:
			log.Println("got a job request")
			go func(job struct{}) {
				log.Println("waiting for open worker")
				jobChannel := <-d.WorkerPool

				log.Println("sending work to idle worker")
				jobChannel <- job
			}(job)
		}
	}
}
