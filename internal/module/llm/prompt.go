package llm

import (
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// buildPromptTemplate constructs the chat template used to generate a quest from
// the given world name, world description, and emotional tone.
func buildPromptTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(schema.FString,
		schema.SystemMessage("你是一个专业的游戏任务设计师。你需要根据给定的世界设定和感情基调，生成一个符合游戏逻辑的冒险任务。\n\n设计原则：\n1. 任务名称要简洁有力，体现任务核心目标\n2. 任务描述要生动，融入世界观背景和感情基调\n3. 步骤之间要有逻辑递进关系，形成完整的任务链条\n4. NPC的设计要与世界设定相符，台词要体现角色性格和感情基调\n5. NPC的坐标位置要合理分布，不要全部集中在同一位置\n6. 步骤中的targetNpcId必须对应npcs数组中已定义的NPC的id\n7. x和y的最大值是8\n8. step的type当前仅支持talkToNpc\n9. 确保所有对话台词与感情基调一致"),
		schema.UserMessage("请根据以下信息生成一个冒险任务。\n\n世界名称：{worldname}\n世界观描述：{worlddesc}\n感情基调：{emotion}"),
	)
}
