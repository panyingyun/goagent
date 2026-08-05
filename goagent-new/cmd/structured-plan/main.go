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

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/structured-plan <goal>")
		os.Exit(2)
	}
	client := llm.NewHTTPClient(os.Getenv("LLM_API_KEY"), envOr("LLM_BASE_URL", "https://api.deepseek.com"), os.Getenv("LLM_MODEL"))
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	prompt := `把目标拆成 1 到 5 步，只输出一个 JSON 对象，不要代码围栏。格式：{"title":"...","steps":[{"id":"s1","task":"..."}]}`
	response, err := client.Complete(ctx, llm.Request{Messages: []llm.Message{
		{Role: llm.RoleSystem, Content: prompt},
		{Role: llm.RoleUser, Content: strings.Join(os.Args[1:], " ")},
	}})
	if err != nil {
		fatal(err)
	}
	plan, err := structured.DecodeTaskPlan(response.Message.Content)
	if err != nil {
		fatal(fmt.Errorf("模型输出未通过校验: %w", err))
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

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

