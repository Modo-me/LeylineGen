package main

import (
	"log"
	"os"

	"quest_generator/internal/asynq"
	"quest_generator/internal/database"
	"quest_generator/internal/module/task"
	"quest_generator/internal/router"
)

func redisAddr() string {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return addr
}

func server_init() {
	db := database.DB_INIT()

	// ---- asynq setup ----
	addr := redisAddr()

	// Client — used to enqueue tasks
	asynqClient := asynq.NewClient(addr)

	// Worker — processes tasks in the background
	srv := asynq.NewWorker(addr)
	go func() {
		log.Printf("Starting asynq worker (Redis: %s)", addr)
		asynq.StartWorker(srv, db)
	}()

	// ---- app modules ----
	taskRepository := task.NewTaskRepository(db)
	taskService := task.NewTaskService(taskRepository, asynqClient)
	taskHandler := task.NewTaskHandler(taskService)

	handlers := &router.Handlers{
		TaskHandler: taskHandler,
	}
	routers := router.SetUpRouters(handlers)
	panic(routers.Run(":8080"))
}
