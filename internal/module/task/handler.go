package task

import "C"
import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService *TaskService
}

func NewTaskHandler(taskService *TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

func (t *TaskHandler) CreateTask(c *gin.Context) {
	var taskInfo TaskInfo
	if err := c.ShouldBindJSON(&taskInfo); err != nil {
		C.JSON(400, gin.H{"error": "Invalid request"})
	}
	taskId, err := t.taskService.CreateTask(&taskInfo)
	if err != nil {
		C.JSON(500, gin.H{"error": "Failed to add task"})
	}
	c.JSON(201, gin.H{"taskId": taskId, "status": "pending"})
}

func (t *TaskHandler) QueryTaskState(c *gin.Context) {
	taskIdStr := c.Param("id")
	taskId, err := strconv.ParseUint(taskIdStr, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid taskId"})
	}
	stateResp, err := t.taskService.QueryTaskState(uint(taskId))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to query task result"})
	}
	c.JSON(200, stateResp)
}
