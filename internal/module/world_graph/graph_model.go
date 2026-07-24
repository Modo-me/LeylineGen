package world_graph

type Relationship interface {
	Type() string
	Properties() map[string]any
}

// direction represents a compass direction stored on a ConnectedRel.
type direction string

const (
	North     direction = "north"
	Northeast direction = "northeast"
	East      direction = "east"
	Southeast direction = "southeast"
	South     direction = "south"
	Southwest direction = "southwest"
	West      direction = "west"
	Northwest direction = "northwest"
)

type ConnectedRel struct {
	Direction direction
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

type PlayerNode struct {
	X int64 `json:"x"`
	Z int64 `json:"z"`
}
