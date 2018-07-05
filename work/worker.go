package work

import "log"

// Worker will accept jobs and process them
type Worker struct {
	WorkerPool chan chan struct{}
	JobChannel chan struct{}
	quit       chan bool
}

// NewWorker will create a new worker to be added to the pool
func NewWorker(workerPool chan chan struct{}) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan struct{}),
		quit:       make(chan bool),
	}
}

// Start method starts the run loop for the worker,
// listening for a quit signal in case we need to stop it
func (w Worker) Start() {
	go func() {
		for {
			// Register the current worker into the worker queue
			w.WorkerPool <- w.JobChannel

			select {
			case <-w.JobChannel:
				// we got a request to do some work
				log.Println("finished work")
			case <-w.quit:
				log.Println("exiting worker")
				// we got a quit signal
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests
func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}
