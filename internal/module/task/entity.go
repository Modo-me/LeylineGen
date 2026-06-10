package task

type Task struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	WorldName string     `gorm:"type:varchar(200)" json:"world_name"`
	WorldDesc string     `gorm:"type:text" json:"world_desc"`
	Emotion   string     `gorm:"type:varchar(100)" json:"emotion"`
	State     string     `gorm:"type:varchar(20)" json:"state"`
	Result    TaskResult `gorm:"serializer:json"`
}

type TaskResult struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Steps       []QuestStep `json:"steps"`
	Npcs        []Npc       `json:"npcs"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Npc struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Position      Position `json:"position"`
	DialogueLines []string `json:"dialogueLines"`
}

type QuestStep struct {
	Type        string `json:"type"`
	TargetNpcId string `json:"targetNpcId"`
}
