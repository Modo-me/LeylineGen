package task

// TaskInfo dto data between frontend and handler
type TaskInfo struct {
	WorldName string
	WorldDesc string
	Emotion   string
}

// TODO: add percentage field
type StateRespInfo struct {
	state string
}
