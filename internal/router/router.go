package router

import (
	"quest_generator/internal/module/task"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	TaskHandler *task.TaskHandler
}

func SetUpRouters(h *Handlers) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.POST("/tasks", h.TaskHandler.CreateTask)
		api.GET("/tasks/:id", h.TaskHandler.QueryTaskResult)
	}
	return router
}
