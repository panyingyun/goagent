package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"goagent-new/pkg/llm"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/llm-once <question>")
		os.Exit(2)
	}
	client := llm.NewHTTPClientFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	response, err := client.Complete(ctx, llm.Request{Messages: []llm.Message{{Role: llm.RoleUser, Content: strings.Join(os.Args[1:], " ")}}})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(response.Message.Content)
	fmt.Fprintf(os.Stderr, "usage: prompt=%d completion=%d total=%d request_id=%s\n", response.Usage.PromptTokens, response.Usage.CompletionTokens, response.Usage.TotalTokens, response.RequestID)
}
