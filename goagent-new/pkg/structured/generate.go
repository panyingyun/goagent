package structured

import (
	"context"
	"fmt"
	"strings"

	"goagent-new/pkg/llm"
)

// MaxAttempts 是结构化输出的最大尝试次数(含首次)。失败后会把校验错误反馈给模型再修。
const MaxAttempts = 3

const taskPlanSystemPrompt = `你是任务规划器。根据用户目标输出一个 JSON 对象，不要 Markdown，不要解释。

硬性要求：
1. 只输出一个 JSON 对象，必须以 { 开头、以 } 结尾；
2. 字段只能是 title 与 steps，禁止其它字段；
3. title 为非空字符串；
4. steps 为数组，长度 1 到 5；
5. 每个 step 含非空 id、task；id 在数组内唯一，建议 s1、s2…；
6. 即使用户目标像计算题，也要输出“如何计算/验证”的步骤计划，禁止只输出最终数字或纯推演文本。

示例：
{"title":"发布 Agent 服务","steps":[{"id":"s1","task":"补齐健康检查"},{"id":"s2","task":"配置鉴权"}]}`

// CreateTaskPlan 使用 JSON mode + temperature=0，并在校验失败时有限次修复重试。
func CreateTaskPlan(ctx context.Context, client llm.Client, goal string) (TaskPlan, error) {
	if client == nil {
		return TaskPlan{}, fmt.Errorf("llm client is required")
	}
	goal = strings.TrimSpace(goal)
	if goal == "" {
		return TaskPlan{}, fmt.Errorf("goal is required")
	}

	temperature := 0.0
	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: taskPlanSystemPrompt},
		{Role: llm.RoleUser, Content: goal},
	}

	var lastRaw string
	var lastErr error
	for attempt := 1; attempt <= MaxAttempts; attempt++ {
		fmt.Println("attempt----->", attempt)
		if err := ctx.Err(); err != nil {
			return TaskPlan{}, err
		}
		response, err := client.Complete(ctx, llm.Request{
			Messages:       messages,
			Temperature:    &temperature,
			ResponseFormat: &llm.ResponseFormat{Type: "json_object"},
		})
		if err != nil {
			return TaskPlan{}, fmt.Errorf("structured output attempt %d: %w", attempt, err)
		}
		lastRaw = response.Message.Content
		plan, err := DecodeTaskPlan(lastRaw)
		if err == nil {
			return plan, nil
		}
		lastErr = err
		messages = append(messages,
			llm.Message{Role: llm.RoleAssistant, Content: lastRaw},
			llm.Message{Role: llm.RoleUser, Content: fmt.Sprintf(
				"校验失败: %v\n请重新只输出一个合法 JSON 对象，字段仅为 title 与 steps(1~5 步)，不要解释。", err,
			)},
		)
	}
	return TaskPlan{}, fmt.Errorf("structured output failed after %d attempts: %w; last_raw=%q", MaxAttempts, lastErr, truncateRunes(lastRaw, 240))
}

func truncateRunes(text string, max int) string {
	runes := []rune(text)
	if len(runes) <= max {
		return text
	}
	return string(runes[:max]) + "..."
}
