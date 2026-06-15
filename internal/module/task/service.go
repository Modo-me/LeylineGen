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

func NewTaskService(taskRepository *TaskRepository, taskQueue TaskQueue) *TaskService {
	return &TaskService{
		taskRepository: taskRepository,
		taskQueue:      taskQueue,
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

func (ts *TaskService) QueryTaskResult(ctx context.Context, id uint) (ResultRespInfo, error) {
	task, err := ts.taskRepository.QueryTask(ctx, id)
	if err != nil {
		return ResultRespInfo{}, err
	}
	resultInfo := ResultRespInfo{
		State:  task.State,
		Result: task.Result,
	}
	return resultInfo, err
}

func (ts *TaskService) QueryTaskInfo(ctx context.Context, id uint) (*TaskInfo, error) {
	task, err := ts.taskRepository.QueryTask(ctx, id)
	if err != nil {
		return nil, err
	}

	taskInfo := &TaskInfo{
		WorldName: task.WorldName,
		WorldDesc: task.WorldDesc,
		Emotion:   task.Emotion,
	}
	return taskInfo, nil
}

func (ts *TaskService) UpdateTask(ctx context.Context, task *Task) error {
	return ts.taskRepository.UpdateTask(ctx, task)
}
