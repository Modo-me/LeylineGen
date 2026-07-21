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
		api.GET("/tasks/:id", h.TaskHandler.QueryTaskResult)
		api.GET("/quest/result", h.QuestHandler.GetResult)
	}
	return router
}
