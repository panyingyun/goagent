package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"goagent-new/pkg/agent"
	"goagent-new/pkg/llm"
)

func runAgent(ctx context.Context, client llm.Client, question string) (string, []llm.Message, error) {
	executor := agent.NewExecutor()
	if err := executor.Register(agent.NewCalculatorTool()); err != nil {
		return "", nil, err
	}
	runner := agent.Runner{LLM: client, Executor: executor, MaxRounds: 5}
	return runner.Run(ctx, []llm.Message{
		{Role: llm.RoleSystem, Content: "需要计算时必须调用 calculator，不要心算。"},
		{Role: llm.RoleUser, Content: question},
	})
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./ch03-tool-calling <question>")
		os.Exit(2)
	}
	client := llm.NewHTTPClientFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	answer, history, err := runAgent(ctx, client, strings.Join(os.Args[1:], " "))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, message := range history {
		if message.Role == llm.RoleTool {
			fmt.Fprintf(os.Stderr, "tool[%s] => %s\n", message.ToolCallID, message.Content)
		}
	}
	fmt.Println(answer)
}

