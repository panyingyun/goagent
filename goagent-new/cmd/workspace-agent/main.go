package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"goagent-new/pkg/agent"
	"goagent-new/pkg/llm"
	"goagent-new/pkg/trace"
	"goagent-new/pkg/workspace"
)

func main() {
	client := llm.NewHTTPClientFromEnv()
	root := envOr("WORKSPACE_DIR", filepath.FromSlash("testdata/workspace"))
	reader := bufio.NewReader(os.Stdin)
	approver := agent.ApproverFunc(func(ctx context.Context, req agent.ApprovalRequest) (bool, error) {
		fmt.Printf("\n需要批准：工具=%s 副作用=%s\n参数=%s\n允许执行？[y/N] ", req.ToolName, req.SideEffect, req.Arguments)
		answer, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}
		if err := ctx.Err(); err != nil {
			return false, err
		}
		return strings.EqualFold(strings.TrimSpace(answer), "y"), nil
	})
	recorder := trace.NewRecorder()
	app, err := workspace.New(root, client, approver, recorder)
	if err != nil {
		fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	fmt.Println("Workspace Agent 已启动。输入 /exit 退出。")
	var history []llm.Message
	for {
		fmt.Print("\n你> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println()
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "/exit" {
			break
		}
		answer, updated, err := app.Chat(ctx, history, line)
		if err != nil {
			fmt.Fprintln(os.Stderr, "错误:", err)
			continue
		}
		history = trimHistory(updated, 24)
		fmt.Println("Agent>", answer)
	}
	if os.Getenv("AGENT_TRACE") == "1" {
		_ = recorder.Export(os.Stderr)
	}
}

func trimHistory(messages []llm.Message, max int) []llm.Message {
	if len(messages) <= max {
		return messages
	}
	result := make([]llm.Message, 0, max)
	start := len(messages) - max
	if messages[0].Role == llm.RoleSystem {
		result = append(result, messages[0])
		start++
	}
	// 从完整的 user turn 开始，避免留下孤立的 tool result 或 tool_calls。
	for start < len(messages) && messages[start].Role != llm.RoleUser {
		start++
	}
	if start < len(messages) {
		result = append(result, messages[start:]...)
	}
	return result
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
