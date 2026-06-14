package producer

import (
	"encoding/json"
	"log"
	"quest_generator/internal/asynq/queue_common"

	"github.com/hibiken/asynq"
)

type Producer struct {
	client *asynq.Client
}

func NewProducer(addr string) *Producer {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: addr})
	return &Producer{client}
}

func (p *Producer) EnqueueTask(taskID uint) error {
	payload, err := json.Marshal(queue_common.TaskProcessPayload{TaskID: taskID})
	if err != nil {
		return err
	}
	newTask := asynq.NewTask(queue_common.TypeTaskProcess, payload)
	taskInfo, err := p.client.Enqueue(newTask)
	if err != nil {
		log.Printf("asynq: Failed to enqueue asynq newTask %d: %v", taskID, err)
	} else {
		log.Printf("asynq: Enqueued asynq newTask %d (asynq ID: %v)", taskID, taskInfo.ID)
	}
	return err
}
