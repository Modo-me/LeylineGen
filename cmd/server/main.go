package main

import (
	"context"
	"quest_generator/internal/database/graph"
	"quest_generator/internal/database/relational"
	"quest_generator/internal/module/task"
	"quest_generator/internal/router"
)

func main() {
	ctx := context.Background()

	// Initialize graph database
	driver := graph.DbInit()
	defer driver.Close(ctx)
	if err := driver.VerifyConnectivity(ctx); err != nil {
		panic(err)
	}

	// Initialize relational database
	db := relational.DbInit()

	// Initialize queue producer
	producer, redisAddr := producerInit()

	taskRepository := task.NewTaskRepository(db)
	taskService := task.NewTaskService(taskRepository, producer)
	taskHandler := task.NewTaskHandler(taskService)

	// Initialize queue consumer
	consumerInit(redisAddr, taskService)

	// Set up routers
	handlers := &router.Handlers{
		TaskHandler: taskHandler,
	}
	routers := router.SetUpRouters(handlers)

	panic(routers.Run(":8080"))
}
