package main

import (
	"log"
	"quest_generator/internal/asynq/consumer"
	"quest_generator/internal/asynq/producer"

	"quest_generator/internal/module/task"
)

func consumerInit(addr string, taskService *task.TaskService) {
	worker := consumer.NewWorker(addr)
	taskMux := consumer.NewProcessHandler(taskService)
	go func() {
		log.Printf("Starting asynq (Redis: %s)", addr)
		consumer.StartWorker(worker, taskMux)
	}()
}

func producerInit(addr string) *producer.Producer {
	newProducer := producer.NewProducer(addr)
	return newProducer
}
