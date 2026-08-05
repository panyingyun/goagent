package main

import (
	"context"
	"fmt"
	"os"
)

// Agent 是一个有"角色"的独立智能体,能记住自己的对话历史。
type Agent struct {
	Name   string
	Role   string // 系统提示词,决定这个 Agent 的职责
	LLM    ChatLLM
	Memory []Message
}

func (a *Agent) Run(ctx context.Context, task string) (string, error) {
	messages := []Message{{Role: "system", Content: a.Role}, {Role: "user", Content: task}}
	reply, err := a.LLM.Chat(ctx, messages)
	if err != nil {
		return "", err
	}
	a.Memory = append(a.Memory, Message{Role: "user", Content: task}, reply)
	return reply.Content, nil
}

// Orchestra 负责多 Agent 编排:规划者 -> 执行者 -> 审查者,最多 N 轮。
type Orchestra struct {
	Planner   *Agent
	Executor  *Agent
	Reviewer  *Agent
	MaxRounds int
}

func (o *Orchestra) Run(ctx context.Context, task string) (string, error) {
	plan, err := o.Planner.Run(ctx, task)
	if err != nil {
		return "", err
	}
	logf("[规划者] %s\n", plan)

	result := ""
	for round := 1; round <= o.MaxRounds; round++ {
		logf("--- 第 %d 轮 ---\n", round)
		result, err = o.Executor.Run(ctx, "计划: "+plan+"\n请输出执行结果。")
		if err != nil {
			return "", err
		}
		logf("[执行者] %s\n", truncateStr(result, 80))
		review, err := o.Reviewer.Run(ctx, "计划: "+plan+"\n结果: "+result+"\n请审查,若需改进请说明,若通过请回复\"通过\"。")
		if err != nil {
			return "", err
		}
		logf("[审查者] %s\n", truncateStr(review, 80))
		if contains(review, "通过") {
			return result, nil
		}
	}
	return result, nil
}

// logf 把编排日志打到 stderr。
func logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func contains(s, sub string) bool { return len(s) > 0 && indexOf(s, sub) >= 0 }

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func truncateStr(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "..."
}
