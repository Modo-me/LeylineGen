package world_graph

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"gorm.io/gorm"
)

type Repository struct {
	driver neo4j.Driver
	db     *gorm.DB
}

func NewRepository(driver neo4j.Driver, db *gorm.DB) *Repository {
	return &Repository{driver: driver, db: db}
}

func safeString(record *neo4j.Record, key string) string {
	val, ok := record.Get(key)
	if !ok || val == nil {
		return ""
	}
	s, _ := val.(string)
	return s
}

func safeInt64(record *neo4j.Record, key string) int64 {
	val, ok := record.Get(key)
	if !ok || val == nil {
		return 0
	}
	i, _ := val.(int64)
	return i
}

// GetConnectedVillages returns all VillageNodes directly connected to the
// given node via a Connected relationship.
func (r *Repository) GetConnectedVillages(ctx context.Context, node VillageNode) ([]VillageNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v:Village {Id: $id})-[r:Connected]->(connected:Village)
		 RETURN connected.ID AS Id, connected.Name AS Name, connected.X AS X, connected.Z AS Z`,
		map[string]any{"id": node.ID},
	)
	if err != nil {
		return nil, fmt.Errorf("get connected villages: %w", err)
	}

	var nodes []VillageNode
	for result.Next(ctx) {
		record := result.Record()
		nodes = append(nodes, VillageNode{
			ID:   safeInt64(record, "Id"),
			Name: safeString(record, "Name"),
			X:    safeInt64(record, "X"),
			Z:    safeInt64(record, "Z"),
		})
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate connected villages: %w", err)
	}
	return nodes, nil
}

// GetConnectedVillageByDirection returns the VillageNode connected to the
// given node via a Connected relationship whose properties match the
// provided rel.
func (r *Repository) GetConnectedVillageByDirection(ctx context.Context, node VillageNode, rel ConnectedRel) (*VillageNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v:Village {Id: $id})-[r:Connected {Direction: $dir}]->(connected:Village)
		 RETURN connected.ID AS Id, connected.Name AS Name, connected.X AS X, connected.Z AS Z`,
		map[string]any{"id": node.ID, "dir": string(rel.Direction)},
	)
	if err != nil {
		return nil, fmt.Errorf("get connected village: %w", err)
	}

	if !result.Next(ctx) {
		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("iterate connected village: %w", err)
		}
		return nil, nil
	}

	record := result.Record()
	return &VillageNode{
		ID:   safeInt64(record, "Id"),
		Name: safeString(record, "Name"),
		X:    safeInt64(record, "X"),
		Z:    safeInt64(record, "Z"),
	}, nil
}

// GetPlayerConnectedVillages returns all VillageNodes connected to the
// player via a Connected relationship.
func (r *Repository) GetPlayerConnectedVillages(ctx context.Context) ([]VillageNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (p:player)-[:Connected]->(v:Village)
		 RETURN v.ID AS Id, v.Name AS Name, v.X AS X, v.Z AS Z`,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("get player connected villages: %w", err)
	}

	var nodes []VillageNode
	for result.Next(ctx) {
		record := result.Record()
		nodes = append(nodes, VillageNode{
			ID:   safeInt64(record, "Id"),
			Name: safeString(record, "Name"),
			X:    safeInt64(record, "X"),
			Z:    safeInt64(record, "Z"),
		})
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate player villages: %w", err)
	}
	return nodes, nil
}

// GetAllNpcsInVillage returns all NpcNodes connected to the given VillageNode
// via a HasNpc relationship.
func (r *Repository) GetAllNpcsInVillage(ctx context.Context, node VillageNode) ([]NpcNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v:Village {Id: $id})-[:HasNpc]->(n:Npc)
		 RETURN n.ID AS Id, n.Name AS Name`,
		map[string]any{"id": node.ID},
	)
	if err != nil {
		return nil, fmt.Errorf("get npcs: %w", err)
	}

	var nodes []NpcNode
	for result.Next(ctx) {
		record := result.Record()
		nodes = append(nodes, NpcNode{
			ID:   safeInt64(record, "Id"),
			Name: safeString(record, "Name"),
		})
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate npcs: %w", err)
	}
	return nodes, nil
}

// GetNpcNodeByID returns the NpcNode with the given ID.
func (r *Repository) GetNpcNodeByID(ctx context.Context, id int64) (*NpcNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (n:Npc {Id: $id})
		 RETURN n.ID AS Id, n.Name AS Name`,
		map[string]any{"id": id},
	)
	if err != nil {
		return nil, fmt.Errorf("get npc node: %w", err)
	}

	if !result.Next(ctx) {
		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("iterate npc node: %w", err)
		}
		return nil, nil
	}

	record := result.Record()
	return &NpcNode{
		ID:   safeInt64(record, "Id"),
		Name: safeString(record, "Name"),
	}, nil
}

// GetVillageByNpc returns the VillageNode that hosts the given NpcNode via a
// HasNpc relationship.
func (r *Repository) GetVillageByNpc(ctx context.Context, node NpcNode) (*VillageNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v:Village)-[:HasNpc]->(n:Npc {Id: $id})
		 RETURN v.ID AS Id, v.Name AS Name, v.X AS X, v.Z AS Z`,
		map[string]any{"id": node.ID},
	)
	if err != nil {
		return nil, fmt.Errorf("get village: %w", err)
	}

	if !result.Next(ctx) {
		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("iterate village: %w", err)
		}
		return nil, nil
	}

	record := result.Record()
	return &VillageNode{
		ID:   safeInt64(record, "Id"),
		Name: safeString(record, "Name"),
		X:    safeInt64(record, "X"),
		Z:    safeInt64(record, "Z"),
	}, nil
}

// GetConnectedRel returns the ConnectedRel between two VillageNodes,
// matching from source toward target.
func (r *Repository) GetConnectedRel(ctx context.Context, source, target VillageNode) (*ConnectedRel, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v1:Village {Id: $sourceId}), (v2:Village {Id: $targetId})
		 MATCH (v1)-[r:Connected]->(v2)
		 RETURN r.Direction AS Direction
		 LIMIT 1`,
		map[string]any{"sourceId": source.ID, "targetId": target.ID},
	)
	if err != nil {
		return nil, fmt.Errorf("get connected rel: %w", err)
	}

	if !result.Next(ctx) {
		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("iterate connected rel: %w", err)
		}
		return nil, nil
	}

	record := result.Record()
	return &ConnectedRel{
		Direction: direction(safeString(record, "Direction")),
	}, nil
}

// CreateNpcNodeByVillage creates a Npc node in Neo4j and links it to the
// specified Village via a HasNpc relationship.
func (r *Repository) CreateNpcNodeByVillage(ctx context.Context, node *NpcNode, villageID int64) error {
	if node == nil {
		return fmt.Errorf("npc node is nil")
	}

	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v:Village {Id: $villageId})
		 CREATE (v)-[:HasNpc]->(n:Npc {Id: $npcId, Name: $npcName})
		 RETURN n`,
		map[string]any{
			"villageId": villageID,
			"npcId":     node.ID,
			"npcName":   node.Name,
		},
	)
	if err != nil {
		return fmt.Errorf("create npc node by village: %w", err)
	}

	summary, err := result.Consume(ctx)
	if err != nil {
		return fmt.Errorf("consume npc result: %w", err)
	}
	if summary.Counters().NodesCreated() == 0 {
		return fmt.Errorf("village %d not found in Neo4j: MATCH returned no nodes", villageID)
	}
	return nil
}

func (r *Repository) CreateVillageNode(ctx context.Context, node *VillageNode) error {
	if node == nil {
		return fmt.Errorf("village node is nil")
	}

	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.Run(ctx,
		`CREATE (v:Village {Id: $id, Name: $name, X: $x, Z: $z})
		 RETURN v`,
		map[string]any{
			"id":   node.ID,
			"name": node.Name,
			"x":    node.X,
			"z":    node.Z,
		},
	)
	if err != nil {
		return fmt.Errorf("create village node: %w", err)
	}
	return nil
}

func (r *Repository) CreateConnectedRel(ctx context.Context, source, target *VillageNode, rel ConnectedRel) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.Run(ctx,
		`MATCH (v1:Village {Id: $sourceId}), (v2:Village {Id: $targetId})
		 CREATE (v1)-[:Connected {Direction: $direction}]->(v2)
		 RETURN v1, v2`,
		map[string]any{
			"sourceId":  source.ID,
			"targetId":  target.ID,
			"direction": string(rel.Direction),
		},
	)
	if err != nil {
		return fmt.Errorf("create connected rel: %w", err)
	}
	return nil
}

// CreatePlayerNode creates or updates the singleton player node with the
// given X, Z coordinates.
func (r *Repository) CreatePlayerNode(ctx context.Context, node *PlayerNode) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.Run(ctx,
		`MERGE (p:player)
		 SET p.X = $x, p.Z = $z`,
		map[string]any{"x": node.X, "z": node.Z},
	)
	if err != nil {
		return fmt.Errorf("create player node: %w", err)
	}
	return nil
}

// QueryAllVillages returns every Village node in the graph.
func (r *Repository) QueryAllVillages(ctx context.Context) ([]VillageNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v:Village)
		 RETURN v.ID AS Id, v.Name AS Name, v.X AS X, v.Z AS Z`,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("query all villages: %w", err)
	}

	var nodes []VillageNode
	for result.Next(ctx) {
		record := result.Record()
		nodes = append(nodes, VillageNode{
			ID:   safeInt64(record, "Id"),
			Name: safeString(record, "Name"),
			X:    safeInt64(record, "X"),
			Z:    safeInt64(record, "Z"),
		})
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate villages: %w", err)
	}
	return nodes, nil
}

// QueryPlayerNode returns the singleton player node with its coordinates.
func (r *Repository) QueryPlayerNode(ctx context.Context) (*PlayerNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (p:player)
		 RETURN p.X AS X, p.Z AS Z`,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("query player node: %w", err)
	}

	if !result.Next(ctx) {
		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("iterate player node: %w", err)
		}
		return nil, nil
	}

	record := result.Record()
	return &PlayerNode{
		X: safeInt64(record, "X"),
		Z: safeInt64(record, "Z"),
	}, nil
}

// CreateVillageToVillageRel creates a directed Connected relationship from
// source village to target village with the given compass direction.
func (r *Repository) CreateVillageToVillageRel(ctx context.Context, sourceID, targetID int64, dir direction) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v1:Village {Id: $sid}), (v2:Village {Id: $tid})
		 CREATE (v1)-[:Connected {Direction: $dir}]->(v2)`,
		map[string]any{"sid": sourceID, "tid": targetID, "dir": string(dir)},
	)
	if err != nil {
		return fmt.Errorf("create village→village rel: %w", err)
	}

	summary, err := result.Consume(ctx)
	if err != nil {
		return fmt.Errorf("consume village-village rel: %w", err)
	}
	if summary.Counters().RelationshipsCreated() == 0 {
		return fmt.Errorf("village %d or %d not found in Neo4j", sourceID, targetID)
	}
	return nil
}

// CreatePlayerToVillageRel creates a directed Connected relationship from
// the player node to the given village.
func (r *Repository) CreatePlayerToVillageRel(ctx context.Context, villageID int64, dir direction) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (p:player), (v:Village {Id: $vid})
		 CREATE (p)-[:Connected {Direction: $dir}]->(v)`,
		map[string]any{"vid": villageID, "dir": string(dir)},
	)
	if err != nil {
		return fmt.Errorf("create player→village rel: %w", err)
	}

	summary, err := result.Consume(ctx)
	if err != nil {
		return fmt.Errorf("consume player-village rel: %w", err)
	}
	if summary.Counters().RelationshipsCreated() == 0 {
		return fmt.Errorf("player node or village %d not found in Neo4j", villageID)
	}
	return nil
}

// CreateVillageToPlayerRel creates a directed Connected relationship from
// the given village to the player node.
func (r *Repository) CreateVillageToPlayerRel(ctx context.Context, villageID int64, dir direction) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v:Village {Id: $vid}), (p:player)
		 CREATE (v)-[:Connected {Direction: $dir}]->(p)`,
		map[string]any{"vid": villageID, "dir": string(dir)},
	)
	if err != nil {
		return fmt.Errorf("create village→player rel: %w", err)
	}

	summary, err := result.Consume(ctx)
	if err != nil {
		return fmt.Errorf("consume village-player rel: %w", err)
	}
	if summary.Counters().RelationshipsCreated() == 0 {
		return fmt.Errorf("village %d or player node not found in Neo4j", villageID)
	}
	return nil
}
