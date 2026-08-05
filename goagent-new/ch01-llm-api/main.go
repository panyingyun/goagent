package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"goagent-new/pkg/llm"
)

func ask(ctx context.Context, client llm.Client, question string) (llm.Response, error) {
	return client.Complete(ctx, llm.Request{Messages: []llm.Message{
		{Role: llm.RoleSystem, Content: "你是简洁、诚实的中文助手。"},
		{Role: llm.RoleUser, Content: question},
	}})
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./ch01-llm-api <question>")
		os.Exit(2)
	}
	client := llm.NewHTTPClient(os.Getenv("LLM_API_KEY"), envOr("LLM_BASE_URL", "https://api.deepseek.com"), os.Getenv("LLM_MODEL"))
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	response, err := ask(ctx, client, strings.Join(os.Args[1:], " "))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(response.Message.Content)
	fmt.Fprintf(os.Stderr, "finish=%s prompt_tokens=%d completion_tokens=%d request_id=%s\n",
		response.FinishReason, response.Usage.PromptTokens, response.Usage.CompletionTokens, response.RequestID)
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

