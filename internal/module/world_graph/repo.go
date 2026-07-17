package world_graph

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type Repository struct {
	driver neo4j.Driver
}

func NewRepository(driver neo4j.Driver) *Repository {
	return &Repository{driver: driver}
}

func safeString(record *neo4j.Record, key string) string {
	val, ok := record.Get(key)
	if !ok || val == nil {
		return ""
	}
	s, _ := val.(string)
	return s
}

// GetConnectedVillages returns all VillageNodes directly connected to the
// given node via a Connected relationship.
func (r *Repository) GetConnectedVillages(ctx context.Context, node VillageNode) ([]VillageNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v:Village {Id: $id})-[r:Connected]->(connected:Village)
		 RETURN connected.Id AS Id, connected.Name AS Name`,
		map[string]any{"id": node.Id},
	)
	if err != nil {
		return nil, fmt.Errorf("get connected villages: %w", err)
	}

	var nodes []VillageNode
	for result.Next(ctx) {
		record := result.Record()
		nodes = append(nodes, VillageNode{
			Id:   safeString(record, "Id"),
			Name: safeString(record, "Name"),
		})
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate connected villages: %w", err)
	}
	return nodes, nil
}

// GetConnectedVillageByDirection returns the VillageNode connected to the given node
// via a Connected relationship whose properties match the provided rel.
func (r *Repository) GetConnectedVillageByDirection(ctx context.Context, node VillageNode, rel ConnectedRel) (*VillageNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v:Village {Id: $id})-[r:Connected {Direction: $dir}]->(connected:Village)
		 RETURN connected.Id AS Id, connected.Name AS Name`,
		map[string]any{"id": node.Id, "dir": rel.Direction},
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
		Id:   safeString(record, "Id"),
		Name: safeString(record, "Name"),
	}, nil
}

// GetPlayerConnectedVillages returns all VillageNodes connected to the
// player via a Connected relationship.
func (r *Repository) GetPlayerConnectedVillages(ctx context.Context) ([]VillageNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (p:player)-[:Connected]->(v:Village)
		 RETURN v.Id AS Id, v.Name AS Name`,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("get player connected villages: %w", err)
	}

	var nodes []VillageNode
	for result.Next(ctx) {
		record := result.Record()
		nodes = append(nodes, VillageNode{
			Id:   safeString(record, "Id"),
			Name: safeString(record, "Name"),
		})
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate player villages: %w", err)
	}
	return nodes, nil
}

// GetAllNpcsInVillage returns all NpcNodes connected to the given VillageNode via a
// HasNpc relationship.
func (r *Repository) GetAllNpcsInVillage(ctx context.Context, node VillageNode) ([]NpcNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v:Village {Id: $id})-[:HasNpc]->(n:Npc)
		 RETURN n.Id AS Id`,
		map[string]any{"id": node.Id},
	)
	if err != nil {
		return nil, fmt.Errorf("get npcs: %w", err)
	}

	var nodes []NpcNode
	for result.Next(ctx) {
		record := result.Record()
		nodes = append(nodes, NpcNode{Id: safeString(record, "Id")})
	}
	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("iterate npcs: %w", err)
	}
	return nodes, nil
}

// GetVillageByNpc returns the VillageNode that hosts the given NpcNode via a
// HasNpc relationship.
func (r *Repository) GetVillageByNpc(ctx context.Context, node NpcNode) (*VillageNode, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v:Village)-[:HasNpc]->(n:Npc {Id: $id})
		 RETURN v.Id AS Id, v.Name AS Name`,
		map[string]any{"id": node.Id},
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
		Id:   safeString(record, "Id"),
		Name: safeString(record, "Name"),
	}, nil
}

// GetConnectedRel returns the ConnectedRel between two VillageNodes.
// It matches regardless of the relationship's arrow direction.
func (r *Repository) GetConnectedRel(ctx context.Context, source, target VillageNode) (*ConnectedRel, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		`MATCH (v1:Village {Id: $sourceId}), (v2:Village {Id: $targetId})
		 MATCH (v1)-[r:Connected]->(v2)
		 RETURN r.Direction AS Direction
		 LIMIT 1`,
		map[string]any{"sourceId": source.Id, "targetId": target.Id},
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
		Direction: safeString(record, "Direction"),
	}, nil
}
