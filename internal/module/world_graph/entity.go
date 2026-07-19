package world_graph

type Relationship interface {
	Type() string
	Properties() map[string]any
}

type ConnectedRel struct {
	Direction string
}

func (c ConnectedRel) Type() string {
	return "Connected"
}

func (c ConnectedRel) Properties() map[string]any {
	return map[string]any{
		"Direction": c.Direction,
	}
}

type HasNpcRel struct {
}

func (h HasNpcRel) Type() string {
	return "HasNpc"
}

func (h HasNpcRel) Properties() map[string]any {
	return map[string]any{}
}

type NpcNode struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type VillageNode struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	X    int64  `json:"x"`
	Z    int64  `json:"z"`
}

type Npc struct {
	ID int64 `gorm:"primaryKey" json:"id"`
}

type Step struct {
	ID            int64    `gorm:"primaryKey" json:"id"`
	QuestID       int64    `gorm:"index" json:"quest_id"`
	NpcID         int64    `gorm:"index" json:"npc_id"`
	DialogueLines []string `gorm:"serializer:json" json:"dialogue_lines"`
	Npc           Npc      `gorm:"foreignKey:NpcID" json:"npc"`
}

type Quest struct {
	ID    int64  `gorm:"primaryKey" json:"id"`
	Steps []Step `gorm:"foreignKey:QuestID" json:"steps"`
}
