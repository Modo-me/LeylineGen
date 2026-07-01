package llm

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/spf13/viper"

	"quest_generator/internal/module/task"
)

var (
	chain     compose.Runnable[map[string]any, *schema.Message] // 干净 Prompt + Tool 绑定
	chainJSON compose.Runnable[map[string]any, *schema.Message] // JSON 约束 Prompt（回落用）
	questTool *schema.ToolInfo
)

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

	// ① 干净 Prompt 链：无 JSON 约束，模型如支持 tool calling 则自然走 ToolCalls
	c1 := compose.NewChain[map[string]any, *schema.Message]()
	c1.AppendChatTemplate(buildCleanPromptTemplate())
	c1.AppendChatModel(chatModel)
	chain, err = c1.Compile(ctx)
	if err != nil {
		panic("failed to compile clean chain: " + err.Error())
	}

	// ② JSON Prompt 链：带结构约束，作为 reasoning 模型的回落
	c2 := compose.NewChain[map[string]any, *schema.Message]()
	c2.AppendChatTemplate(buildJSONPromptTemplate())
	c2.AppendChatModel(chatModel)
	chainJSON, err = c2.Compile(ctx)
	if err != nil {
		panic("failed to compile JSON fallback chain: " + err.Error())
	}

	questTool = defineToolInfo()

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

	// ① 快速路径：干净 Prompt + 工具 → 模型如支持 tool calling 则直接调用
	msg, err := chain.Invoke(ctx, input,
		compose.WithChatModelOption(
			model.WithTools([]*schema.ToolInfo{questTool}),
			model.WithToolChoice(schema.ToolChoiceAllowed),
		),
	)
	if err != nil {
		return nil, err
	}
	if len(msg.ToolCalls) > 0 {
		log.Printf("successfully invoked task with tool calls: %s", msg.ToolCalls[0].Function.Name)
		tc := msg.ToolCalls[0]
		var result task.TaskResult
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &result); err != nil {
			log.Printf("tool call JSON parse failed: %v\nraw arguments:\n%s", err, tc.Function.Arguments)
		} else {
			b, _ := json.MarshalIndent(result, "", "  ")
			log.Printf("raw json:\n%s", string(b))
			return &result, nil
		}
	}

	// ② 回落路径：模型不调工具（如 reasoning 模型）→ JSON 约束 Prompt + Content 解析
	log.Printf("tool call path failed, falling back to JSON-constrained prompt")
	msg, err = chainJSON.Invoke(ctx, input)
	if err != nil {
		return nil, err
	}

	var result task.TaskResult
	if err := json.Unmarshal([]byte(msg.Content), &result); err != nil {
		return nil, errors.New("model returned non-JSON: " + msg.Content)
	}

	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Printf("marshal error: %v", err)
		return nil, err
	}
	log.Printf("raw json:\n%s", string(b))
	return &result, nil
}
