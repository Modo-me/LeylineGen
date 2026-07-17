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
	Id string `json:"id"`
}

type VillageNode struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
