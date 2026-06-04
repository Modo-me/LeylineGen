package task

type Task struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	WorldName string     `gorm:"type:varchar(200)" json:"world_name"`
	WorldDesc string     `gorm:"type:text" json:"world_desc"`
	Emotion   string     `gorm:"type:varchar(100)" json:"emotion"`
	State     string     `gorm:"type:varchar(20)" json:"state"`
	Result    taskResult `gorm:"serializer:json"`
}

type taskResult struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Steps       []questStep `json:"steps"`
	Npcs        []npc       `json:"npcs"`
}

type position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type npc struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Position      position `json:"position"`
	DialogueLines []string `json:"dialogueLines"`
}

type questStep struct {
	Type        string `json:"type"`
	TargetNpcId string `json:"targetNpcId"`
}
