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
		fmt.Fprintln(os.Stderr, "usage: go run ./ch02-structured-output <goal>")
		os.Exit(2)
	}
	client := llm.NewHTTPClientFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	plan, err := structured.CreateTaskPlan(ctx, client, strings.Join(os.Args[1:], " "))
	if err != nil {
		fmt.Fprintln(os.Stderr, "结构化输出失败:", err)
		os.Exit(1)
	}
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
