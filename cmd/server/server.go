package main

import (
	"os"

	"quest_generator/internal/database"
	"quest_generator/internal/module/task"
	"quest_generator/internal/router"
)

func serverInit() {
	db := database.DB_INIT()

	addr := redisAddr()

	asyncQueue := asynqInit(addr)

	taskRepository := task.NewTaskRepository(db)
	taskService := task.NewTaskService(taskRepository, asyncQueue)
	taskHandler := task.NewTaskHandler(taskService)

	handlers := &router.Handlers{
		TaskHandler: taskHandler,
	}
	routers := router.SetUpRouters(handlers)
	panic(routers.Run(":8080"))
}

func redisAddr() string {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	return addr
}
