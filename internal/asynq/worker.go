package asynq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// NewWorker creates an asynq server (worker) connected to the given Redis address.
func NewWorker(redisAddr string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 10},
	)
}

// StartWorker registers handlers and starts the worker server.
func StartWorker(srv *asynq.Server, db *gorm.DB) {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeTaskProcess, func(ctx context.Context, t *asynq.Task) error {
		var payload TaskProcessPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return err
		}
		log.Printf("Processing task %d", payload.TaskID)

		// Update the task state to PROCESSING
		if err := db.Table("tasks").
			Where("id = ?", payload.TaskID).
			Update("state", "PROCESSING").Error; err != nil {
			return err
		}

		// Placeholder: actual quest generation logic goes here.
		// For now, simulate a quest result and mark as COMPLETED.
		result := map[string]interface{}{
			"id":          "quest-001",
			"name":        "Generated Quest",
			"description": "A quest generated from world context",
			"steps":       []map[string]interface{}{},
			"npcs":        []map[string]interface{}{},
		}
		resultJSON, _ := json.Marshal(result)

		if err := db.Exec(
			"UPDATE tasks SET state = ?, result = ? WHERE id = ?",
			"COMPLETED", string(resultJSON), payload.TaskID,
		).Error; err != nil {
			return err
		}

		log.Printf("Task %d completed", payload.TaskID)
		return nil
	})

	if err := srv.Start(mux); err != nil {
		log.Fatalf("asynq worker failed: %v", err)
	}
}
