package world_graph

// ==== Result data to be returned to the client ====

type Result struct {
	Steps []StepResult `json:"steps"`
}
type NpcResult struct {
	Name string `json:"name"`
	X    int64  `json:"x"`
	Z    int64  `json:"z"`
}

type StepResult struct {
	NpcResult     NpcResult `json:"npcResult"`
	DialogueLines []string  `json:"dialogueLines"`
}

// ==== Tool calling data to be returned for agent =====

type NearbyWorldInfo struct {
	VillagesInfo []VillageInfo `json:"villagesInfo"`
}

type VillageInfo struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Direction string `json:"direction"`
}

// ==== Client Request data to be sent to the service ====

type VillageCreationRequest struct {
	X int64 `json:"x"`
	Z int64 `json:"z"`
}

type PlayerCreationRequest struct {
	X int64 `json:"x"`
	Z int64 `json:"z"`
}
