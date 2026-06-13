package main

import (
	"log"
	"quest_generator/internal/asynq"
)

func asynqInit(addr string) *asynq.AsyncQueue {
	// AsyncQueue — enqueues tasks to be processed in the background
	asyncQueue := asynq.NewAsyncQueue(addr)

	// Worker — processes tasks in the background
	srv := asynq.NewWorker(addr)
	taskMux := asynq.NewProcessHandler()
	go func() {
		log.Printf("Starting asynq worker (Redis: %s)", addr)
		asynq.StartWorker(srv, taskMux)
	}()
	return asyncQueue
}
