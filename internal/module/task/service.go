package task

import (
	"context"
	"log"
)

type TaskService struct {
	taskRepository *TaskRepository
	taskQueue      TaskQueue
}

type TaskQueue interface {
	EnqueueTask(taskID uint) error
}

func NewTaskService(taskRepository *TaskRepository, taskQue TaskQueue) *TaskService {
	return &TaskService{
		taskRepository: taskRepository,
		taskQueue:      taskQue,
	}
}

func (ts *TaskService) CreateTask(ctx context.Context, taskInfo *TaskInfo) (uint, error) {
	task := Task{
		WorldName: taskInfo.WorldName,
		WorldDesc: taskInfo.WorldDesc,
		Emotion:   taskInfo.Emotion,
		State:     "PENDING",
	}
	taskID, err := ts.taskRepository.CreateTask(ctx, &task)
	if err != nil {
		return 0, err
	}

	// Enqueue async processing job for this task
	if ts.taskQueue != nil {
		err := ts.taskQueue.EnqueueTask(taskID)
		if err != nil {
			log.Printf("failed to enqueue task %d for async processing: %v", taskID, err)
			return 0, err
		}
	}

	return taskID, nil
}

func (ts *TaskService) QueryTaskState(ctx context.Context, id uint) (StateRespInfo, error) {
	taskState, err := ts.taskRepository.QueryTaskState(ctx, id)
	respInfo := StateRespInfo{
		state: taskState,
	}
	return respInfo, err
}
