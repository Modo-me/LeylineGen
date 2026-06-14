package consumer

import (
	"context"
	"encoding/json"
	"log"
	"quest_generator/internal/asynq/queue_common"
	"quest_generator/internal/module/task"

	"github.com/hibiken/asynq"
)

// StartWorker starts the producer server with the given ServeMux.
// The caller is responsible for registering handlers on mux before calling this.
func StartWorker(srv *asynq.Server, mux *asynq.ServeMux) {
	if err := srv.Start(mux); err != nil {
		log.Fatalf("asynq producer failed: %v", err)
	}
}

// NewWorker creates an asynq server (producer) connected to the given Redis address.
func NewWorker(redisAddr string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 10},
	)
}

// NewProcessHandler returns a ServeMux with the task:process handler registered.
// The handler reads the task from the queue, processes it, and writes the result.
func NewProcessHandler(taskService *task.TaskService) *asynq.ServeMux {
	tp := &taskProcessor{taskService: taskService}
	mux := asynq.NewServeMux()
	mux.HandleFunc(queue_common.TypeTaskProcess, tp.processTask)
	return mux
}

type taskProcessor struct {
	taskService *task.TaskService
}

func (tp *taskProcessor) processTask(ctx context.Context, t *asynq.Task) error {
	var payload queue_common.TaskProcessPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	taskId := payload.TaskID
	log.Printf("Processing task %d", taskId)
	return nil
}
