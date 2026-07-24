package router

import (
	"quest_generator/internal/module/task"
	"quest_generator/internal/module/world_graph"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	TaskHandler  *task.TaskHandler
	QuestHandler *world_graph.Handler
}

func SetUpRouters(h *Handlers) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.POST("/tasks", h.TaskHandler.CreateTask)
		api.GET("/quest/result", h.QuestHandler.GetResult)
		api.POST("/village", h.QuestHandler.CreateVillage)
	}
	return router
}
