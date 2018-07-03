package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/deciphernow/gm-fabric-go/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
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
	JobQueue = make(chan struct{}, MaxQueue)

	// Start our dispatcher
	dispatcher := NewDispatcher(MaxWorkers)
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
	)

	s := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: stack.Wrap(mux),
	}

	s.ListenAndServe()
}

func handler(w http.ResponseWriter, r *http.Request) {
	work := Job{}

	log.Println("Sending work to job queue")
	JobQueue <- work

	w.WriteHeader(http.StatusOK)
}
