package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Chain 是 LangChain 的核心抽象:一次"输入 -> 输出"的变换。
type Chain interface {
	Call(ctx context.Context, input string) (string, error)
}

// LLMChain 最基础的链:系统提示词 + 输入 -> LLM 回复。
type LLMChain struct {
	LLM    ChatLLM
	System string
}

func (c *LLMChain) Call(ctx context.Context, input string) (string, error) {
	reply, err := c.LLM.Chat(ctx, []Message{{Role: "system", Content: c.System}, {Role: "user", Content: input}})
	if err != nil {
		return "", err
	}
	return reply.Content, nil
}

// SequentialChain 把多条链串起来:上一条的输出是下一条的输入。
type SequentialChain struct {
	Chains []Chain
}

func (c *SequentialChain) Call(ctx context.Context, input string) (string, error) {
	out := input
	for i, ch := range c.Chains {
		var err error
		out, err = ch.Call(ctx, out)
		if err != nil {
			return "", fmt.Errorf("第 %d 条链失败: %w", i, err)
		}
	}
	return out, nil
}

// Memory 是简单的对话记忆(对应 LangChain 的 Memory 概念)。
type Memory struct {
	Messages []Message
}

func (m *Memory) Add(role, content string) {
	m.Messages = append(m.Messages, Message{Role: role, Content: content})
}

// Tool 对应 LangChain 的 Tool。
type Tool struct {
	Name        string
	Description string
	Run         func(ctx context.Context, input string) (string, error)
}

// AgentExecutor 对应 LangChain 的 AgentExecutor:LLM + 工具 + 循环。
type AgentExecutor struct {
	LLM      ChatLLM
	Tools    []Tool
	MaxSteps int
}

func (a *AgentExecutor) Call(ctx context.Context, input string) (string, error) {
	system := "你是一个 Agent。需要工具时输出:\nAction: 工具名\nAction Input: {json}\n完成后输出:\nFinal Answer: 答案\n可用工具:\n"
	for _, t := range a.Tools {
		system += "- " + t.Name + ": " + t.Description + "\n"
	}
	messages := []Message{{Role: "system", Content: system}, {Role: "user", Content: input}}
	for step := 0; step < a.MaxSteps; step++ {
		reply, err := a.LLM.Chat(ctx, messages)
		if err != nil {
			return "", err
		}
		messages = append(messages, reply)
		content := reply.Content
		if m := finalAnswerRe.FindStringSubmatch(content); m != nil {
			return strings.TrimSpace(m[1]), nil
		}
		m := actionRe.FindStringSubmatch(content)
		if m == nil {
			return "", fmt.Errorf("格式错误: %s", content)
		}
		action := m[1]
		jsonPart := extractJSONObject(content)
		obs, err := a.runTool(ctx, action, jsonPart)
		if err != nil {
			obs = "错误: " + err.Error()
		}
		messages = append(messages, Message{Role: "user", Content: "Observation: " + obs})
	}
	return "", fmt.Errorf("超过最大步数")
}

func (a *AgentExecutor) runTool(ctx context.Context, name, jsonArgs string) (string, error) {
	for _, t := range a.Tools {
		if t.Name == name {
			return t.Run(ctx, jsonArgs)
		}
	}
	return "", fmt.Errorf("未知工具: %s", name)
}

var finalAnswerRe = regexp.MustCompile(`(?s)Final Answer:\s*(.+)`)
var actionRe = regexp.MustCompile(`(?s)Action:\s*([A-Za-z_][A-Za-z0-9_]*)\s*`)

func extractJSONObject(s string) string {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return ""
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}
