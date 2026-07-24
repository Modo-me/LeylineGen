package world_graph

import (
	"context"
	"fmt"
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
	quest := &Quest{
		Steps: steps,
	}
	err := s.Repository.CreateQuest(ctx, quest)
	if err != nil {
		return nil
	}

	for _, step := range steps {
		step.QuestID = quest.ID
		err := s.Repository.UpdateStep(ctx, &step)
		if err != nil {
			return nil
		}
	}
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
