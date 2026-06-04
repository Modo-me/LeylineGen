package task

// TaskInfo dto data between frontend and handler
type TaskInfo struct {
	worldName string
	worldDesc string
	emotion   string
}

// TODO: add percentage field
type StateRespInfo struct {
	state string
}
