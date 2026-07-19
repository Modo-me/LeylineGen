package world_graph

type NpcResult struct {
	Name string `json:"name"`
	X    int64  `json:"x"`
	Z    int64  `json:"z"`
}

type StepResult struct {
	NpcResult     NpcResult `json:"npcResult"`
	DialogueLines []string  `json:"dialogueLines"`
}

type Result struct {
	Steps []StepResult `json:"steps"`
}
