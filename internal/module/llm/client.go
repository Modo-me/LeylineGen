package llm

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"

	"quest_generator/internal/module/task"
)

var llmRunner compose.Runnable[map[string]any, *schema.Message]

func initLLMConfig() *viper.Viper {
	v := viper.New()

	v.SetDefault("llm.base_url", "https://api.openai.com/v1")
	v.SetDefault("llm.model", "gpt-4o")
	v.SetDefault("llm.api_key", "")

	v.SetConfigName("llm")
	v.SetConfigType("yaml")
	v.AddConfigPath("internal/module/llm")
	v.AddConfigPath(".")
	_ = v.ReadInConfig()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	return v
}

func init() {
	ctx := context.Background()
	cfg := initLLMConfig()

	baseURL := cfg.GetString("llm.base_url")
	apiKey := cfg.GetString("llm.api_key")
	if apiKey == "" {
		panic("llm.api_key is required: set it in internal/module/llm/llm.yaml or LLM_API_KEY env var")
	}
	modelName := cfg.GetString("llm.model")

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   modelName,
	})
	if err != nil {
		panic("failed to create chat model: " + err.Error())
	}

	chain := compose.NewChain[map[string]any, *schema.Message]()
	chain.AppendChatTemplate(buildPromptTemplate())
	chain.AppendChatModel(chatModel)

	runner, err := chain.Compile(ctx)
	if err != nil {
		panic("failed to compile LLM chain: " + err.Error())
	}
	llmRunner = runner
	log.Printf("successfully initialized LLM runner with model %s at %s", modelName, baseURL)
}

// ProcessTask generates a quest task result based on the given world context.
//
// Parameters:
//   - worldname: the name of the world to generate a quest for
//   - worlddesc: a description of the world's setting and lore
//   - emotion:   the emotional tone for the quest (e.g. "dark", "hopeful", "mysterious")
//
// Returns a TaskResult containing the generated quest details, or an error.
func ProcessTask(worldname, worlddesc, emotion string) (*task.TaskResult, error) {
	ctx := context.Background()

	input := map[string]any{
		"worldname": worldname,
		"worlddesc": worlddesc,
		"emotion":   emotion,
	}

	msg, err := llmRunner.Invoke(ctx, input)
	if err != nil {
		return nil, err
	}

	var result task.TaskResult
	if err := json.Unmarshal([]byte(msg.Content), &result); err != nil {
		return nil, err
	}
	return &result, nil
}
