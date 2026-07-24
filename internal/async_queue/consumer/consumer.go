package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"quest_generator/internal/async_queue/queue_common"
	"quest_generator/internal/module/agent"
	"quest_generator/internal/module/task"
	"quest_generator/internal/module/world_graph"

	"github.com/hibiken/asynq"
)

func StartWorker(srv *asynq.Server, mux *asynq.ServeMux) {
	if err := srv.Start(mux); err != nil {
		log.Fatalf("async_queue producer failed: %v", err)
	}
}

func NewWorker(redisAddr string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 10},
	)
}

func NewProcessHandler(taskService *task.TaskService, wgSvc *world_graph.Service) *asynq.ServeMux {
	tp := &taskProcessor{taskService: taskService, wgSvc: wgSvc}
	mux := asynq.NewServeMux()
	mux.HandleFunc(queue_common.TypeTaskProcess, tp.processTask)
	return mux
}

type taskProcessor struct {
	taskService *task.TaskService
	wgSvc       *world_graph.Service
}

func (tp *taskProcessor) processTask(ctx context.Context, t *asynq.Task) error {
	var payload queue_common.TaskProcessPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	taskId := payload.TaskID
	log.Printf("Processing task %d", taskId)

	taskInfo, err := tp.taskService.QueryTaskInfo(ctx, taskId)
	if err != nil {
		return err
	}

	// Build the world graph connections.
	if tp.wgSvc != nil {
		if err := tp.wgSvc.BuildWorldGraph(ctx); err != nil {
			return fmt.Errorf("build world graph: %w", err)
		}
	}

	if err = agent.ProcessTask(ctx, taskInfo.WorldName, taskInfo.WorldDesc, taskInfo.Emotion); err != nil {
		log.Printf("Failed to process task %d: %v", taskId, err)
		return err
	}

	if err = tp.taskService.UpdateTask(ctx, &task.Task{
		ID:        taskId,
		WorldName: taskInfo.WorldName,
		WorldDesc: taskInfo.WorldDesc,
		Emotion:   taskInfo.Emotion,
		State:     "COMPLETED",
	}); err != nil {
		log.Printf("Failed to update task %d: %v", taskId, err)
		return err
	}

	log.Printf("task %d processed successfully", taskId)
	return nil
}
