package world_graph

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// StepInput carries the data needed to create a single Quest step.
type StepInput struct {
	NpcID         int64
	DialogueLines []string
}

type service struct {
	Repository *Repository
}

func NewService(repo *Repository) *service {
	return &service{Repository: repo}
}

// CreateQuest creates a Quest along with NPC and Step records in a single
// transaction and returns the new quest ID.
func (s *service) CreateQuest(ctx context.Context, steps []StepInput) (int64, error) {
	var quest Quest
	err := s.Repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		seen := make(map[int64]bool)
		for _, step := range steps {
			if seen[step.NpcID] {
				continue
			}
			seen[step.NpcID] = true
			npc := Npc{ID: step.NpcID}
			if err := tx.FirstOrCreate(&npc).Error; err != nil {
				return fmt.Errorf("ensure npc %d: %w", step.NpcID, err)
			}
		}

		if err := tx.Create(&quest).Error; err != nil {
			return fmt.Errorf("create quest: %w", err)
		}

		for _, step := range steps {
			s := Step{
				QuestID:       quest.ID,
				NpcID:         step.NpcID,
				DialogueLines: step.DialogueLines,
			}
			if err := tx.Create(&s).Error; err != nil {
				return fmt.Errorf("create step for npc %d: %w", step.NpcID, err)
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return quest.ID, nil
}

// BuildResult aggregates data from the relational store and world_graph to
// produce the final Result for a given quest.
func (s *service) BuildResult(ctx context.Context, questID int64) (*Result, error) {
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
