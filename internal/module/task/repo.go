package task

import (
	"context"

	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}

func (tr *TaskRepository) CreateTask(ctx context.Context, task *Task) (uint, error) {
	return task.ID, tr.db.WithContext(ctx).Create(task).Error
}

func (tr *TaskRepository) QueryTask(ctx context.Context, taskId uint) (*Task, error) {
	var task Task
	result := tr.db.WithContext(ctx).First(&task, taskId)
	if result.Error != nil {
		return nil, result.Error
	}
	return &task, nil
}

func (tr *TaskRepository) QueryTaskState(ctx context.Context, taskId uint) (string, error) {
	var task Task
	result := tr.db.WithContext(ctx).First(&task, taskId)
	if result.Error != nil {
		return "", result.Error
	}
	return task.State, nil
}

func (tr *TaskRepository) UpdateTask(ctx context.Context, task *Task) error {
	return tr.db.WithContext(ctx).Save(task).Error
}
