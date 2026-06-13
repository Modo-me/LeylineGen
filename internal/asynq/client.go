package asynq

import (
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
)

const TypeTaskProcess = "task:process"

type TaskProcessPayload struct {
	TaskID uint `json:"task_id"`
}

type AsyncQueue struct {
	Client *asynq.Client
}

func NewAsyncQueue(addr string) *AsyncQueue {
	// Create a new asynq client connected to the Redis server.
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: addr})

	asyncQueue := &AsyncQueue{asynqClient}
	return asyncQueue
}

// EnqueueTask enqueues a task-processing job for the given task ID.
func (aq *AsyncQueue) EnqueueTask(taskID uint) error {
	payload, err := json.Marshal(TaskProcessPayload{TaskID: taskID})
	if err != nil {
		return err
	}
	task := asynq.NewTask(TypeTaskProcess, payload)
	taskInfo, err := aq.Client.Enqueue(task)
	if err != nil {
		log.Printf("asynq: Failed to enqueue asynq task %d: %v", taskID, err)
	} else {
		log.Printf("asynq: Enqueued asynq task %d (asynq ID: %v)", taskID, taskInfo.ID)
	}
	return err
}
