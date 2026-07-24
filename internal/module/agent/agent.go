package agent

import (
	"context"
	"fmt"
	"log"
	"os"

	"quest_generator/internal/module/world_graph"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
	"gopkg.in/yaml.v3"
)

var (
	llmCfg    *Config
	globalSvc *world_graph.Service
)

type Config struct {
	Model   string `yaml:"model"`
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
}

// ---- Agent lifecycle ----------------------------------------------------

func InitConfig() {
	data, err := os.ReadFile("internal/module/agent/llm.yaml")
	if err != nil {
		log.Fatalf("read agent config: %v", err)
	}
	llmCfg = &Config{}
	if err := yaml.Unmarshal(data, llmCfg); err != nil {
		log.Fatalf("parse agent config: %v", err)
	}
}

func Init(svc *world_graph.Service) {
	globalSvc = svc
}

// ProcessTask runs a ReAct agent loop that observes the world, creates NPCs
// and dialogue steps, then persists everything via CreateQuestWithSteps.
func ProcessTask(ctx context.Context, worldName, worldDesc, emotion string) error {

	// Wrap context with per-task storage, so concurrent tasks don't
	// interfere with each other's steps.
	ctx = world_graph.NewTaskContext(ctx)
	if llmCfg == nil {
		InitConfig()
	}
	if globalSvc == nil {
		return fmt.Errorf("agent not initialized; call agent.Init() first")
	}

	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   llmCfg.Model,
		APIKey:  llmCfg.APIKey,
		BaseURL: llmCfg.BaseURL,
	})
	if err != nil {
		return fmt.Errorf("create chat model: %w", err)
	}

	tools := newTools(globalSvc)
	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, t := range tools {
		toolInfos = append(toolInfos, t.info)
	}
	if err := cm.BindTools(toolInfos); err != nil {
		return fmt.Errorf("bind tools: %w", err)
	}

	msgs := []*schema.Message{
		{Role: schema.System, Content: SystemPrompt},
		{Role: schema.User, Content: fmt.Sprintf(
			"世界名称：%s\n世界描述：%s\n感情基调：%s",
			worldName, worldDesc, emotion,
		)},
	}

	const maxIter = 15
	for iter := 0; iter < maxIter; iter++ {
		resp, err := cm.Generate(ctx, msgs)
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		// No tool calls → persist the quest (guard runs inside CreateQuestWithSteps).
		if len(resp.ToolCalls) == 0 {
			return globalSvc.CreateQuestWithSteps(ctx)
		}

		msgs = append(msgs, resp)

		for _, tc := range resp.ToolCalls {
			var result string
			for _, t := range tools {
				if t.info.Name == tc.Function.Name {
					result, err = t.execute(ctx, tc.Function.Arguments)
					break
				}
			}
			if err != nil {
				return fmt.Errorf("execute tool %s: %w", tc.Function.Name, err)
			}
			msgs = append(msgs, &schema.Message{
				Role:       schema.Tool,
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}
	return fmt.Errorf("agent exceeded %d iterations", maxIter)
}

