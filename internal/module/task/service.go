package task

type TaskService struct {
	taskRepository *TaskRepository
}

func NewTaskService(taskRepository *TaskRepository) *TaskService {
	return &TaskService{
		taskRepository: taskRepository,
	}
}

func (ts *TaskService) CreateTask(taskInfo *TaskInfo) (uint, error) {
	task := Task{
		WorldName: taskInfo.worldName,
		WorldDesc: taskInfo.worldDesc,
		Emotion:   taskInfo.emotion,
		State:     "PENDING",
	}
	return ts.taskRepository.CreateTask(&task)
}

func (ts *TaskService) QueryTaskState(id uint) (StateRespInfo, error) {
	taskState, err := ts.taskRepository.QueryTaskState(id)
	respInfo := StateRespInfo{
		state: taskState,
	}
	return respInfo, err
}
