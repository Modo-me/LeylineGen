package task

import (
	"context"
	"log"

	myasynq "quest_generator/internal/asynq"

	"github.com/hibiken/asynq"
)

type TaskService struct {
	taskRepository *TaskRepository
	asynqClient    *asynq.Client
}

func NewTaskService(taskRepository *TaskRepository, asynqClient *asynq.Client) *TaskService {
	return &TaskService{
		taskRepository: taskRepository,
		asynqClient:    asynqClient,
	}
}

func (ts *TaskService) CreateTask(ctx context.Context, taskInfo *TaskInfo) (uint, error) {
	task := Task{
		WorldName: taskInfo.worldName,
		WorldDesc: taskInfo.worldDesc,
		Emotion:   taskInfo.emotion,
		State:     "PENDING",
	}
	taskID, err := ts.taskRepository.CreateTask(ctx, &task)
	if err != nil {
		return 0, err
	}

	// Enqueue async processing job via asynq
	if ts.asynqClient != nil {
		info, err := myasynq.EnqueueProcessTask(ts.asynqClient, taskID)
		if err != nil {
			log.Printf("Failed to enqueue asynq task %d: %v", taskID, err)
		} else {
			log.Printf("Enqueued asynq task %d (asynq ID: %s)", taskID, info.ID)
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
