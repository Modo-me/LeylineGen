package world_graph

type Npc struct {
	ID int64 `gorm:"primaryKey" json:"id"`
}

type Step struct {
	ID            int64    `gorm:"primaryKey" json:"id"`
	QuestID       *int64   `gorm:"index" json:"quest_id"`
	NpcID         int64    `gorm:"index" json:"npc_id"`
	DialogueLines []string `gorm:"serializer:json" json:"dialogue_lines"`
	Npc           Npc      `gorm:"foreignKey:NpcID" json:"npc"`
}

type Quest struct {
	ID    int64  `gorm:"primaryKey" json:"id"`
	Steps []Step `gorm:"foreignKey:QuestID" json:"steps"`
}

type Village struct {
	ID int64 `gorm:"primaryKey" json:"id"`
}
