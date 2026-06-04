package main

import (
	"quest_generator/internal/database"
	"quest_generator/internal/module/task"
	"quest_generator/internal/router"
)

func server_init() {
	db := database.DB_INIT()
	taskRepository := task.NewTaskRepository(db)
	taskService := task.NewTaskService(taskRepository)
	taskHandler := task.NewTaskHandler(taskService)
	handlers := &router.Handlers{
		TaskHandler: taskHandler,
	}
	routers := router.SetUpRouters(handlers)
	panic(routers.Run(":8080"))
}
