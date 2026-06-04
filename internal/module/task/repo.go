package task

import "gorm.io/gorm"

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}

func (tr *TaskRepository) CreateTask(task *Task) (uint, error) {
	return task.ID, tr.db.Create(task).Error
}

func (tr *TaskRepository) QueryTaskState(taskId uint) (string, error) {
	var task Task
	result := tr.db.First(&task, taskId)
	if result.Error != nil {
		return "", result.Error
	}
	return task.State, nil
}
