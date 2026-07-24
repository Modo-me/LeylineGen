package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"quest_generator/internal/module/world_graph"

	"github.com/cloudwego/eino/schema"
)

type toolDef struct {
	info    *schema.ToolInfo
	execute func(ctx context.Context, argsJSON string) (string, error)
}

func newTools(svc *world_graph.Service) []*toolDef {
	return []*toolDef{
		{
			info: &schema.ToolInfo{
				Name: "observe_player_nearby_world",
				Desc: "观察玩家周围的所有村庄，获取村庄ID和名称",
			},
			execute: func(ctx context.Context, _ string) (string, error) {
				info, err := svc.ObservePlayerNearbyWorld(ctx)
				if err != nil {
					return "", err
				}
				b, _ := json.Marshal(info)
				return string(b), nil
			},
		},
		{
			info: &schema.ToolInfo{
				Name: "observe_village_nearby_world",
				Desc: "观察指定村庄周围有哪些村庄以及各村庄的相对方向",
				ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
					"village_id": {Type: schema.Integer, Desc: "村庄ID", Required: true},
				}),
			},
			execute: func(ctx context.Context, argsJSON string) (string, error) {
				var args struct{ VillageID int64 `json:"village_id"` }
				if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
					return "", fmt.Errorf("parse args: %w", err)
				}
				info, err := svc.ObserveVillageNearbyWorld(ctx, args.VillageID)
				if err != nil {
					return "", err
				}
				b, _ := json.Marshal(info)
				return string(b), nil
			},
		},
		{
			info: &schema.ToolInfo{
				Name: "create_new_npc",
				Desc: "在指定村庄创建一个新NPC，返回NPC ID",
				ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
					"npc_name":   {Type: schema.String, Desc: "NPC名称", Required: true},
					"village_id": {Type: schema.Integer, Desc: "NPC所在村庄ID", Required: true},
				}),
			},
			execute: func(ctx context.Context, argsJSON string) (string, error) {
				var args struct {
					NpcName   string `json:"npc_name"`
					VillageID int64  `json:"village_id"`
				}
				if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
					return "", fmt.Errorf("parse args: %w", err)
				}
				npcID, err := svc.CreateNewNpc(ctx, args.NpcName, args.VillageID)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf(`{"npc_id":%d,"status":"created"}`, npcID), nil
			},
		},
		{
			info: &schema.ToolInfo{
				Name: "create_step_with_npc",
				Desc: "为指定NPC创建对话步骤（台词列表）",
				ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
					"npc_id":         {Type: schema.Integer, Desc: "NPC ID", Required: true},
					"dialogue_lines": {Type: schema.Array, ElemInfo: &schema.ParameterInfo{Type: schema.String}, Desc: "该NPC的对话台词列表", Required: true},
				}),
			},
			execute: func(ctx context.Context, argsJSON string) (string, error) {
				// LLM sometimes sends malformed JSON (trailing commas) or a single
				// string for dialogue_lines instead of an array.  Tolerate both.
				cleaned := strings.TrimSpace(argsJSON)
				// Remove trailing comma before ] and }
				for strings.Contains(cleaned, ",]") || strings.Contains(cleaned, ",}") {
					cleaned = strings.ReplaceAll(cleaned, ",]", "]")
					cleaned = strings.ReplaceAll(cleaned, ",}", "}")
				}

				var raw struct {
					NpcID         int64           `json:"npc_id"`
					DialogueLines json.RawMessage `json:"dialogue_lines"`
				}
				if err := json.Unmarshal([]byte(cleaned), &raw); err != nil {
					return "", fmt.Errorf("parse args: %w", err)
				}

				var lines []string
				// Try array first, then single string.
				if err := json.Unmarshal(raw.DialogueLines, &lines); err != nil {
					var single string
					if err2 := json.Unmarshal(raw.DialogueLines, &single); err2 != nil {
						return "", fmt.Errorf("dialogue_lines must be a string or []string")
					}
					lines = []string{single}
				}

				if err := svc.CreateStepWithNpc(ctx, raw.NpcID, lines); err != nil {
					return "", err
				}
				return `{"status":"step_created"}`, nil
			},
		},
	}
}
