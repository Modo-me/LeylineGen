package world_graph

import (
	"context"
	"fmt"
	"math"
)

type Service struct {
	Repository *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repository: repo}
}

var (
	questID int64 = 1
	steps   []Step
)

// CreateQuestWithSteps creates a new quest with the given steps and saves it to the repository.
func (s *Service) CreateQuestWithSteps(ctx context.Context) error {
	for i := range steps {
		steps[i].Npc = Npc{} // 清除，Npc 已存在只需引用 NpcID
	}
	quest := &Quest{
		Steps: steps,
	}
	err := s.Repository.CreateQuest(ctx, quest)
	if err != nil {
		return err // 原代码 return nil 是吞错误，改回 err
	}
	steps = nil
	questID = quest.ID
	return nil
}

// BuildResult aggregates data from the relational store and world_graph to
// produce the final Result for a given quest.
func (s *Service) BuildResult(ctx context.Context) (*Result, error) {
	relQuest, err := s.Repository.QueryQuest(ctx, questID)
	if err != nil {
		return nil, fmt.Errorf("query quest %d: %w", questID, err)
	}

	steps := make([]StepResult, 0, len(relQuest.Steps))
	for _, step := range relQuest.Steps {
		npcNode, err := s.Repository.GetNpcNodeByID(ctx, step.NpcID)
		if err != nil {
			return nil, fmt.Errorf("get npc node %d: %w", step.NpcID, err)
		}
		if npcNode == nil {
			return nil, fmt.Errorf("npc node %d not found in graph", step.NpcID)
		}

		villageNode, err := s.Repository.GetVillageByNpc(ctx, *npcNode)
		if err != nil {
			return nil, fmt.Errorf("get village for npc %d: %w", step.NpcID, err)
		}
		if villageNode == nil {
			return nil, fmt.Errorf("village for npc %d not found", step.NpcID)
		}

		steps = append(steps, StepResult{
			NpcResult: NpcResult{
				Name: npcNode.Name,
				X:    villageNode.X,
				Z:    villageNode.Z,
			},
			DialogueLines: step.DialogueLines,
		})
	}

	return &Result{Steps: steps}, nil
}

func (s *Service) CreateVillage(ctx context.Context, req *VillageCreationRequest) error {
	village := &Village{}
	err := s.Repository.CreateVillage(ctx, village)
	if err != nil {
		return err
	}
	villageNode := &VillageNode{
		ID: village.ID,
		X:  req.X,
		Z:  req.Z,
	}
	err = s.Repository.CreateVillageNode(ctx, villageNode)
	return err
}

func (s *Service) CreatePlayer(ctx context.Context, req *PlayerCreationRequest) error {
	return s.Repository.CreatePlayerNode(ctx, &PlayerNode{X: req.X, Z: req.Z})
}

// BuildWorldGraph connects the player to nearby villages and links nearby
// villages to each other with bidirectional Connected relationships.
// The direction property on each relationship is calculated from coordinates.
func (s *Service) BuildWorldGraph(ctx context.Context) error {
	player, err := s.Repository.QueryPlayerNode(ctx)
	if err != nil {
		return fmt.Errorf("query player: %w", err)
	}
	if player == nil {
		return fmt.Errorf("player node not found; create player first")
	}

	villages, err := s.Repository.QueryAllVillages(ctx)
	if err != nil {
		return fmt.Errorf("query villages: %w", err)
	}

	const radius = 1000.0

	// Player ↔ nearby villages (bidirectional)
	for _, v := range villages {
		dist := distance(player.X, player.Z, v.X, v.Z)
		if dist < radius {
			dirPV := calcDirection(player.X, player.Z, v.X, v.Z)
			dirVP := calcDirection(v.X, v.Z, player.X, player.Z)
			if err := s.Repository.CreatePlayerToVillageRel(ctx, v.ID, dirPV); err != nil {
				return fmt.Errorf("player→village %d: %w", v.ID, err)
			}
			if err := s.Repository.CreateVillageToPlayerRel(ctx, v.ID, dirVP); err != nil {
				return fmt.Errorf("village %d→player: %w", v.ID, err)
			}
		}
	}

	// Village ↔ nearby villages (bidirectional, no self-pair)
	for i := 0; i < len(villages); i++ {
		for j := i + 1; j < len(villages); j++ {
			dist := distance(villages[i].X, villages[i].Z, villages[j].X, villages[j].Z)
			if dist < radius {
				dirAB := calcDirection(villages[i].X, villages[i].Z, villages[j].X, villages[j].Z)
				dirBA := calcDirection(villages[j].X, villages[j].Z, villages[i].X, villages[i].Z)
				if err := s.Repository.CreateVillageToVillageRel(ctx, villages[i].ID, villages[j].ID, dirAB); err != nil {
					return fmt.Errorf("village %d→%d: %w", villages[i].ID, villages[j].ID, err)
				}
				if err := s.Repository.CreateVillageToVillageRel(ctx, villages[j].ID, villages[i].ID, dirBA); err != nil {
					return fmt.Errorf("village %d→%d: %w", villages[j].ID, villages[i].ID, err)
				}
			}
		}
	}

	return nil
}

// ---- coordinate helpers ------------------------------------------------

func distance(x1, z1, x2, z2 int64) float64 {
	dx := float64(x2 - x1)
	dz := float64(z2 - z1)
	return math.Sqrt(dx*dx + dz*dz)
}

// calcDirection returns the compass direction from (x1, z1) to (x2, z2).
func calcDirection(x1, z1, x2, z2 int64) direction {
	dx := float64(x2 - x1)
	dz := float64(z2 - z1)
	if dx == 0 && dz == 0 {
		return North
	}

	deg := math.Atan2(dz, dx) * 180 / math.Pi
	if deg < 0 {
		deg += 360
	}

	switch {
	case deg >= 337.5 || deg < 22.5:
		return East
	case deg >= 22.5 && deg < 67.5:
		return Southeast
	case deg >= 67.5 && deg < 112.5:
		return South
	case deg >= 112.5 && deg < 157.5:
		return Southwest
	case deg >= 157.5 && deg < 202.5:
		return West
	case deg >= 202.5 && deg < 247.5:
		return Northwest
	case deg >= 247.5 && deg < 292.5:
		return North
	default:
		return Northeast
	}
}
