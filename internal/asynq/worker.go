package asynq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
)

// NewWorker creates an asynq server (worker) connected to the given Redis address.
func NewWorker(redisAddr string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 10},
	)
}

// StartWorker starts the worker server with the given ServeMux.
// The caller is responsible for registering handlers on mux before calling this.
func StartWorker(srv *asynq.Server, mux *asynq.ServeMux) {
	if err := srv.Start(mux); err != nil {
		log.Fatalf("asynq worker failed: %v", err)
	}
}

// NewProcessHandler returns a ServeMux with the task:process handler registered.
// The handler reads the task from the queue, processes it, and writes the result.
func NewProcessHandler() *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeTaskProcess, processTask)
	return mux
}

func processTask(ctx context.Context, t *asynq.Task) error {
	var payload TaskProcessPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	log.Printf("Processing task %d", payload.TaskID)
	return nil
}
