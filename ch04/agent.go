package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const maxIterations = 10

const reactSystemPrompt = `你是一个运行在 Go Agent 框架中的 AI 助手 DemoAgent。
你通过"思考-行动-观察"循环(ReAct)来回答用户的问题:先思考需要什么信息,
需要时调用工具,看到 Observation 结果后再继续,直到可以给出最终答案。

【输出格式】每次回复只能包含以下两种之一,不要用 markdown 代码块包裹:

1. 需要调用工具时:
Thought: 你的推理
Action: 工具名
Action Input: {"参数名": "值"}

2. 已有答案时:
Thought: 任务已完成
Final Answer: 给用户的最终答案

【重要规则】
- 数学问题必须用 calculator 工具,禁止心算或估算
- 时间问题必须用 get_time 工具
- 工具名与参数名必须和"可用工具"里完全一致
- Action Input 必须是合法 JSON`

const toolSystemPrompt = `你是一个运行在 Go Agent 框架中的 AI 助手 DemoAgent。
你可以调用系统提供的函数(工具)来获得信息或计算结果。当需要工具时,
在回复中声明 tool_calls(由 API 处理);工具执行结果会以 tool 角色的消息
返回给你,你看到结果后继续推理,最后用普通文本给出最终答案。

【重要规则】
- 数学问题必须调用 calculator,禁止心算或估算
- 时间问题必须调用 get_time
- 不要编造工具结果,一切以 Observation 为准`

func buildSystemPrompt(base string, skills []Skill, reg *Registry) string {
	var b strings.Builder
	b.WriteString(base)
	if len(skills) > 0 {
		b.WriteString("\n\n## 你掌握的技能(SKILL)\n")
		b.WriteString("技能是打包好的行为规范,遇到对应场景必须遵守。\n")
		for _, s := range skills {
			b.WriteString("\n### 技能: " + s.Name + "\n")
			b.WriteString(s.Instructions + "\n")
		}
	}
	b.WriteString("\n## 可用工具\n")
	for _, n := range reg.Names() {
		t := reg.tools[n]
		b.WriteString("- " + t.Name + ": " + t.Description + "\n")
	}
	return b.String()
}

// runReAct 是文本版 ReAct 循环:
// LLM 输出 Thought/Action/Action Input -> 我们执行工具 -> 把 Observation 放回对话 -> 继续,
// 直到模型给出 Final Answer 或达到最大轮数。
func runReAct(ctx context.Context, llm *LLMClient, reg *Registry, skills []Skill, userInput string) (string, error) {
	messages := []Message{
		{Role: RoleSystem, Content: buildSystemPrompt(reactSystemPrompt, skills, reg)},
		{Role: RoleUser, Content: userInput},
	}

	for step := 1; step <= maxIterations; step++ {
		fmt.Fprintf(os.Stderr, "\n--- 第 %d 轮思考 ---\n", step)
		reply, err := llm.Chat(ctx, messages, nil)
		if err != nil {
			return "", err
		}
		messages = append(messages, reply)
		content := strings.TrimSpace(reply.Content)
		fmt.Fprintln(os.Stderr, content)

		if final := extractFinalAnswer(content); final != "" {
			return final, nil
		}

		action, input, ok := parseReAct(content)
		if !ok {
			messages = append(messages, Message{
				Role:    RoleUser,
				Content: "格式错误:你既没有给出 Final Answer,也没有给出 Action。请只使用要求的格式。",
			})
			continue
		}

		obs, err := reg.Run(ctx, action, input)
		if err != nil {
			obs = "错误: " + err.Error()
		}
		fmt.Fprintf(os.Stderr, "Observation: %s\n", obs)
		messages = append(messages, Message{Role: RoleUser, Content: "Observation: " + obs})
	}
	return "", errors.New("达到最大迭代次数,任务未完成")
}

// runToolCall 是原生 function calling 模式:
// 工具定义直接发给模型,模型返回结构化 tool_calls,执行结果用 tool 角色消息回填。
func runToolCall(ctx context.Context, llm *LLMClient, reg *Registry, skills []Skill, userInput string) (string, error) {
	messages := []Message{
		{Role: RoleSystem, Content: buildSystemPrompt(toolSystemPrompt, skills, reg)},
		{Role: RoleUser, Content: userInput},
	}

	for step := 1; step <= maxIterations; step++ {
		fmt.Fprintf(os.Stderr, "\n--- 第 %d 轮 ---\n", step)
		reply, err := llm.Chat(ctx, messages, reg.ToolDefs())
		if err != nil {
			return "", err
		}
		messages = append(messages, reply)

		if len(reply.ToolCalls) > 0 {
			for _, tc := range reply.ToolCalls {
				fmt.Fprintf(os.Stderr, "调用工具 %s(%s)\n", tc.Function.Name, tc.Function.Arguments)
				result, err := reg.Run(ctx, tc.Function.Name, tc.Function.Arguments)
				if err != nil {
					result = "错误: " + err.Error()
				}
				fmt.Fprintf(os.Stderr, "Observation: %s\n", result)
				messages = append(messages, Message{
					Role:       RoleTool,
					ToolCallID: tc.ID,
					Name:       tc.Function.Name,
					Content:    result,
				})
			}
			continue
		}

		if strings.TrimSpace(reply.Content) != "" {
			return strings.TrimSpace(reply.Content), nil
		}
	}
	return "", errors.New("达到最大迭代次数,任务未完成")
}

// ---- 从模型文本里解析 ReAct 字段 ----

var finalRe = regexp.MustCompile(`(?s)Final Answer:\s*(.+)`)
var actionRe = regexp.MustCompile(`(?s)Action:\s*([A-Za-z_][A-Za-z0-9_]*)\s*`)

func extractFinalAnswer(content string) string {
	m := finalRe.FindStringSubmatch(content)
	if m == nil {
		return ""
	}
	return strings.TrimSpace(stripCodeFence(m[1]))
}

func parseReAct(content string) (action, input string, ok bool) {
	m := actionRe.FindStringSubmatch(content)
	if m == nil {
		return "", "", false
	}
	action = m[1]
	idx := strings.Index(content, "Action Input:")
	if idx < 0 {
		return action, "", true // 无参数工具
	}
	rest := content[idx+len("Action Input:"):]
	jsonPart, err := extractJSONObject(rest)
	if err != nil {
		return action, strings.TrimSpace(rest), true // 交给工具层解析
	}
	return action, jsonPart, true
}

// extractJSONObject 用花括号配对找到完整 JSON 对象,兼容多行输入。
func extractJSONObject(s string) (string, error) {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return "", errors.New("找不到 {")
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1], nil
			}
		}
	}
	return "", errors.New("括号不匹配")
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if i := strings.Index(s[3:], "\n"); i >= 0 {
			s = s[3+i+1:]
		} else {
			s = s[3:]
		}
	}
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
