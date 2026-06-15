package llm

import (
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// buildPromptTemplate constructs the chat template used to generate a quest from
// the given world name, world description, and emotional tone.
func buildPromptTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个专业的游戏任务设计师。你需要根据给定的世界设定和感情基调，生成一个符合游戏逻辑的冒险任务。\n\n你的输出必须是严格的 JSON 格式，不包含任何额外的文本或解释。\nJSON 结构必须完全符合以下类型定义：\n\n{{\n  \"id\": \"quest_001\",\n  \"name\": \"任务名称\",\n  \"description\": \"任务描述，100-300字\",\n  \"steps\": [\n    {{\n      \"type\": \"talk_to_npc | collect_item | defeat_enemy | go_to_location | use_item | trigger_event\",\n      \"targetNpcId\": \"目标NPC的ID，若步骤不涉及NPC则为空字符串\"\n    }}\n  ],\n  \"npcs\": [\n    {{\n      \"id\": \"npc_001\",\n      \"name\": \"NPC名称\",\n      \"position\": {{\n        \"x\": 500,\n        \"y\": 300\n      }},\n      \"dialogueLines\": [\"台词1\", \"台词2\"]\n    }}\n  ]\n}}\n\n设计原则：\n1. 任务名称要简洁有力，体现任务核心目标\n2. 任务描述要生动，融入世界观背景和感情基调\n3. 步骤之间要有逻辑递进关系，形成完整的任务链条\n4. NPC的设计要与世界设定相符，台词要体现角色性格和感情基调\n5. NPC的坐标位置要合理分布，不要全部集中在同一位置\n6. 步骤中的targetNpcId必须对应npcs数组中已定义的NPC的id\n7. 确保所有对话台词与感情基调一致"),
		schema.UserMessage("请根据以下信息生成一个冒险任务。\n\n世界名称：{worldname}\n世界观描述：{worlddesc}\n感情基调：{emotion}\n\n请直接输出符合上述格式的 JSON，不要包含任何 markdown 代码块标记或其他解释文字。"),
	)
}
