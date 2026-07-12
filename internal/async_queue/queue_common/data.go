package queue_common

const TypeTaskProcess = "task:process"

type TaskProcessPayload struct {
	TaskID uint `json:"task_id"`
}
