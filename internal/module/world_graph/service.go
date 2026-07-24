package world_graph

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

type Service struct {
	Repository *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{Repository: repo}
}

// Package-level atomic used only by the GET /api/quest/result endpoint
// to return the most recently created quest.
var GlobalQuestID int64 = 1

// taskData holds per-task state (steps etc.) stored in the request context.
// This replaces the old global steps/questID to avoid race conditions when
// multiple asynq tasks run concurrently.
type taskCtxKey struct{}

type taskData struct {
	mu    sync.Mutex
	steps []Step
}

func NewTaskContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, taskCtxKey{}, &taskData{})
}

func getTaskData(ctx context.Context) *taskData {
	td, _ := ctx.Value(taskCtxKey{}).(*taskData)
	return td
}

// CreateQuestWithSteps creates a new quest with the given steps and saves it to the repository.
func (s *Service) CreateQuestWithSteps(ctx context.Context) error {
	td := getTaskData(ctx)
	if td == nil {
		return fmt.Errorf("CreateQuestWithSteps: no task data in context")
	}
	td.mu.Lock()
	defer td.mu.Unlock()

	if len(td.steps) == 0 {
		return fmt.Errorf("task completed but no steps were created")
	}

	for i := range td.steps {
		td.steps[i].Npc = Npc{}
	}
	quest := &Quest{
		Steps: td.steps,
	}
	err := s.Repository.CreateQuest(ctx, quest)
	if err != nil {
		return err
	}
	td.steps = nil
	atomic.StoreInt64(&GlobalQuestID, quest.ID)
	return nil
}

// BuildResult aggregates data from the relational store and world_graph to
// produce the final Result for a given quest.
func (s *Service) BuildResult(ctx context.Context) (*Result, error) {
	relQuest, err := s.Repository.QueryQuest(ctx, atomic.LoadInt64(&GlobalQuestID))
	if err != nil {
		return nil, fmt.Errorf("query quest %d: %w", atomic.LoadInt64(&GlobalQuestID), err)
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

		log.Printf("[DEBUG] step.NpcID=%d npcNode.ID=%d npcNode.Name=%s err=%v", step.NpcID, npcNode.ID, npcNode.Name, err)
		villageNode, err := s.Repository.GetVillageByNpc(ctx, *npcNode)
		log.Printf("[DEBUG] villageNode=%+v err=%v", villageNode, err)
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
	log.Printf("[BuildWorldGraph] start")

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
	log.Printf("[BuildWorldGraph] player=(%d,%d) villages=%d", player.X, player.Z, len(villages))

	const radius = 100000.0 // 100 km – much larger than 1000 to avoid false negatives

	// Collect all links first, then create them in a single session.
	links := make([]relLink, 0, len(villages)*4) // rough upper bound

	// Player ↔ nearby villages (bidirectional)
	for _, v := range villages {
		dist := distance(player.X, player.Z, v.X, v.Z)
		if dist < radius {
			dirPV := calcDirection(player.X, player.Z, v.X, v.Z)
			dirVP := calcDirection(v.X, v.Z, player.X, player.Z)
			links = append(links,
				relLink{Src: 0, Dst: v.ID, Dir: dirPV}, // player→village (player = id 0)
				relLink{Src: v.ID, Dst: 0, Dir: dirVP}, // village→player
			)
		}
	}

	// Village ↔ nearby villages (bidirectional, no self-pair)
	for i := 0; i < len(villages); i++ {
		for j := i + 1; j < len(villages); j++ {
			dist := distance(villages[i].X, villages[i].Z, villages[j].X, villages[j].Z)
			if dist < radius {
				dirAB := calcDirection(villages[i].X, villages[i].Z, villages[j].X, villages[j].Z)
				dirBA := calcDirection(villages[j].X, villages[j].Z, villages[i].X, villages[i].Z)
				links = append(links,
					relLink{Src: villages[i].ID, Dst: villages[j].ID, Dir: dirAB},
					relLink{Src: villages[j].ID, Dst: villages[i].ID, Dir: dirBA},
				)
			}
		}
	}

	log.Printf("[BuildWorldGraph] creating %d relationships via single session", len(links))
	if len(links) > 0 {
		// Use a child context with timeout so a hung Neo4j call won't hang forever.
		bulkCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := s.Repository.BulkCreateRels(bulkCtx, links); err != nil {
			return fmt.Errorf("bulk create rels: %w", err)
		}
	}
	log.Printf("[BuildWorldGraph] done")
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
