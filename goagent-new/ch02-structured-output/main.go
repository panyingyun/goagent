package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"goagent-new/pkg/llm"
	"goagent-new/pkg/structured"
)

func createPlan(ctx context.Context, client llm.Client, goal string) (structured.TaskPlan, error) {
	response, err := client.Complete(ctx, llm.Request{Messages: []llm.Message{
		{Role: llm.RoleSystem, Content: `把目标拆成 1 到 5 步，只输出 JSON：{"title":"...","steps":[{"id":"s1","task":"..."}]}`},
		{Role: llm.RoleUser, Content: goal},
	}})
	if err != nil {
		return structured.TaskPlan{}, err
	}
	return structured.DecodeTaskPlan(response.Message.Content)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./ch02-structured-output <goal>")
		os.Exit(2)
	}
	client := llm.NewHTTPClient(os.Getenv("LLM_API_KEY"), envOr("LLM_BASE_URL", "https://api.deepseek.com"), os.Getenv("LLM_MODEL"))
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	plan, err := createPlan(ctx, client, strings.Join(os.Args[1:], " "))
	if err != nil {
		fmt.Fprintln(os.Stderr, "结构化输出校验失败:", err)
		os.Exit(1)
	}
	data, _ := json.MarshalIndent(plan, "", "  ")
	fmt.Println(string(data))
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

