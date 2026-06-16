package llm

import (
	"github.com/cloudwego/eino/schema"
)

// defineToolInfo creates the ToolInfo for the quest generation output structure.
// This tool acts as a structured output contract: the model "calls" it with
// arguments that match our TaskResult schema, and we extract the data from
// msg.ToolCalls[0].Function.Arguments.
func defineToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "generate_quest",
		Desc: "根据世界设定和感情基调生成一个冒险任务。调用此工具输出生成的任务结果。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"id": {
				Type:     schema.String,
				Desc:     "任务唯一标识符，例如 quest_001",
				Required: true,
			},
			"name": {
				Type:     schema.String,
				Desc:     "任务名称，简洁有力，体现任务核心目标",
				Required: true,
			},
			"description": {
				Type:     schema.String,
				Desc:     "任务描述，100-300字，融入世界观背景和感情基调",
				Required: true,
			},
			"steps": {
				Type:     schema.Array,
				Required: true,
				Desc:     "任务步骤列表，步骤之间要有逻辑递进关系，形成完整的任务链条",
				ElemInfo: &schema.ParameterInfo{
					Type: schema.Object,
					SubParams: map[string]*schema.ParameterInfo{
						"type": {
							Type:     schema.String,
							Desc:     "步骤类型，当前仅支持 talkToNpc",
							Required: true,
						},
						"targetNpcId": {
							Type:     schema.String,
							Desc:     "目标NPC的ID，必须对应npcs数组中已定义的NPC的id，若步骤不涉及NPC则为空字符串",
							Required: true,
						},
					},
				},
			},
			"npcs": {
				Type:     schema.Array,
				Required: true,
				Desc:     "NPC列表，NPC的设计要与世界设定相符，台词要体现角色性格和感情基调",
				ElemInfo: &schema.ParameterInfo{
					Type: schema.Object,
					SubParams: map[string]*schema.ParameterInfo{
						"id": {
							Type:     schema.String,
							Desc:     "NPC的唯一标识符",
							Required: true,
						},
						"name": {
							Type:     schema.String,
							Desc:     "NPC名称",
							Required: true,
						},
						"position": {
							Type:     schema.Object,
							Required: true,
							Desc:     "NPC坐标位置，x和y的最大值为8，不要全部集中在同一位置",
							SubParams: map[string]*schema.ParameterInfo{
								"x": {Type: schema.Integer, Desc: "X坐标，最大值8", Required: true},
								"y": {Type: schema.Integer, Desc: "Y坐标，最大值8", Required: true},
							},
						},
						"dialogueLines": {
							Type:     schema.Array,
							Required: true,
							Desc:     "NPC的对话台词列表，确保所有对话台词与感情基调一致",
							ElemInfo: &schema.ParameterInfo{Type: schema.String},
						},
					},
				},
			},
		}),
	}
}
