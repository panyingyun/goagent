package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

const maxIterations = 10

// runToolCallParallel 是原生 function calling 循环,支持一次并行调用多个工具。
func runToolCallParallel(ctx context.Context, llm *LLMClient, ex *Executor, userInput string) (string, error) {
	messages := []Message{
		{Role: "system", Content: "你是一个 Go Agent。需要信息时调用工具;可以一次声明多个 tool_calls 并行执行。"},
		{Role: "user", Content: userInput},
	}
	for step := 1; step <= maxIterations; step++ {
		fmt.Fprintf(os.Stderr, "\n--- 第 %d 轮 ---\n", step)
		reply, err := llm.Chat(ctx, messages, toolDefs(ex.reg))
		if err != nil {
			return "", err
		}
		messages = append(messages, reply)
		if len(reply.ToolCalls) > 0 {
			results := ex.ExecuteParallel(ctx, reply.ToolCalls)
			for _, r := range results {
				if r.Err != nil {
					fmt.Fprintf(os.Stderr, "%s 失败: %v\n", r.Name, r.Err)
				} else {
					fmt.Fprintf(os.Stderr, "%s -> %s\n", r.Name, r.Out)
				}
				messages = append(messages, Message{
					Role: "tool", ToolCallID: r.ID, Name: r.Name,
					Content: func() string {
						if r.Err != nil {
							return "错误: " + r.Err.Error()
						}
						return r.Out
					}(),
				})
			}
			continue
		}
		if strings.TrimSpace(reply.Content) != "" {
			return strings.TrimSpace(reply.Content), nil
		}
	}
	return "", errors.New("达到最大迭代次数")
}

func toolDefs(reg *Registry) []ToolDef {
	defs := make([]ToolDef, 0, len(reg.tools))
	for _, n := range reg.Names() {
		t, _ := reg.Get(n)
		defs = append(defs, ToolDef{
			Type: "function",
			Function: ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		})
	}
	return defs
}
