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
你可以调用工具来获取信息或执行操作。

【输出格式】每次回复只能包含以下两种之一,不要用 markdown 代码块包裹:

1. 需要调用工具时:
Thought: 你的推理
Action: 工具名
Action Input: {"参数名": "值"}

2. 已有答案时:
Thought: 任务已完成
Final Answer: 给用户的最终答案

【重要规则】
- 数学问题必须用 calculator
- 时间问题必须用 get_time
- 用户要"联网/查资料"时用 http_get
- 用户要"读/写文件"时用 read_file / write_file
- 命令执行只能用 run_command,且只能执行白名单命令
- 不要编造工具结果,一切以 Observation 为准`

func buildSystemPrompt(reg *Registry) string {
	var b strings.Builder
	b.WriteString(reactSystemPrompt)
	b.WriteString("\n## 可用工具\n")
	for _, n := range reg.Names() {
		t := reg.tools[n]
		b.WriteString("- " + t.Name + ": " + t.Description + "\n")
	}
	return b.String()
}

func runReAct(ctx context.Context, llm *LLMClient, reg *Registry, userInput string) (string, error) {
	messages := []Message{
		{Role: RoleSystem, Content: buildSystemPrompt(reg)},
		{Role: RoleUser, Content: userInput},
	}
	for step := 1; step <= maxIterations; step++ {
		fmt.Fprintf(os.Stderr, "\n--- 第 %d 轮思考 ---\n", step)
		reply, err := llm.Chat(ctx, messages)
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
			messages = append(messages, Message{Role: RoleUser, Content: "格式错误:请只使用要求的格式。"})
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
		return action, "", true
	}
	rest := content[idx+len("Action Input:"):]
	jsonPart, err := extractJSONObject(rest)
	if err != nil {
		return action, strings.TrimSpace(rest), true
	}
	return action, jsonPart, true
}

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
