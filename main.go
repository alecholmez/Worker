package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alecholmez/workerPool/dispatch"
	"github.com/deciphernow/gm-fabric-go/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// Job is a work request to be sent into the worker pool
type Job struct{}

type key int

const (
	queueKey key = iota
)

var (
	// MaxWorkers is the maximum number of workers in the worker pool available to complete jobs
	MaxWorkers = 3
	// MaxQueue is the maximum number of jobs the queue can hold
	MaxQueue = 100
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Initialize our job queue
	jobQueue := make(chan struct{}, MaxQueue)

	// Start our dispatcher
	dispatcher := dispatch.NewDispatcher(MaxWorkers, jobQueue)
	dispatcher.Run()

	mux := mux.NewRouter()
	mux.HandleFunc("/", handler)

	stack := middleware.Chain(
		middleware.MiddlewareFunc(hlog.NewHandler(logger)),
		middleware.MiddlewareFunc(hlog.AccessHandler(func(r *http.Request, status int, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Str("path", r.URL.String()).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("Access")
		})),
		middleware.MiddlewareFunc(hlog.UserAgentHandler("user_agent")),
		middleware.Middleware(WithJobQueue(jobQueue)),
	)

	s := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: stack.Wrap(mux),
	}

	s.ListenAndServe()
}

// WithJobQueue ...
func WithJobQueue(queue chan struct{}) middleware.Middleware {
	return middleware.MiddlewareFunc(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), queueKey, queue))
			next.ServeHTTP(w, r)
		})
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	work := Job{}

	log.Println("Sending work to job queue")
	SendToQueue(r, work)

	w.WriteHeader(http.StatusOK)
}

// SendToQueue ...
func SendToQueue(r *http.Request, work struct{}) error {
	queue, ok := r.Context().Value(queueKey).(chan struct{})
	if !ok {
		return errors.New("failed to extract queue from context")
	}

	// Send our work through to the job queue
	queue <- work

	return nil
}
