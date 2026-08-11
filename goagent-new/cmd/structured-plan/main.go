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
	client := llm.NewHTTPClientFromEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	plan, err := structured.CreateTaskPlan(ctx, client, strings.Join(os.Args[1:], " "))
	if err != nil {
		fatal(err)
	}
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		fatal(err)
	}
	fmt.Println(string(data))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
