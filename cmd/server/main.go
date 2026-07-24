package main

import (
	"context"
	"quest_generator/internal/database/graph"
	"quest_generator/internal/database/relational"
	"quest_generator/internal/module/agent"
	"quest_generator/internal/module/task"
	"quest_generator/internal/module/world_graph"
	"quest_generator/internal/router"
)

func main() {
	ctx := context.Background()

	driver := graph.DbInit()
	defer driver.Close(ctx)
	if err := driver.VerifyConnectivity(ctx); err != nil {
		panic(err)
	}

	db := relational.DbInit()

	producer, redisAddr := producerInit()

	taskRepository := task.NewTaskRepository(db)
	taskService := task.NewTaskService(taskRepository, producer)
	taskHandler := task.NewTaskHandler(taskService)

	wgRepo := world_graph.NewRepository(driver, db)
	wgSvc := world_graph.NewService(wgRepo)
	agent.Init(wgSvc)
	wgHandler := world_graph.NewHandler(wgSvc)

	consumerInit(redisAddr, taskService)

	handlers := &router.Handlers{
		TaskHandler:  taskHandler,
		QuestHandler: wgHandler,
	}
	routers := router.SetUpRouters(handlers)

	panic(routers.Run(":8080"))
}
