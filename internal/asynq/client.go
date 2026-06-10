package asynq

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const TypeTaskProcess = "task:process"

type TaskProcessPayload struct {
	TaskID uint `json:"task_id"`
}

// NewClient creates an asynq client connected to the given Redis address.
func NewClient(redisAddr string) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
}

// EnqueueProcessTask enqueues a task-processing job for the given task ID.
func EnqueueProcessTask(client *asynq.Client, taskID uint) (*asynq.TaskInfo, error) {
	payload, err := json.Marshal(TaskProcessPayload{TaskID: taskID})
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(TypeTaskProcess, payload)
	return client.Enqueue(task)
}
