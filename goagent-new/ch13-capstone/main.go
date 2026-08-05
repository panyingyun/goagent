package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"goagent-new/pkg/agent"
	"goagent-new/pkg/llm"
	"goagent-new/pkg/trace"
	"goagent-new/pkg/workspace"
)

func runCapstone(ctx context.Context, client llm.Client, root, question string, approveWrites bool) (string, *trace.Recorder, error) {
	recorder := trace.NewRecorder()
	approver := agent.ApproverFunc(func(context.Context, agent.ApprovalRequest) (bool, error) {
		return approveWrites, nil
	})
	app, err := workspace.New(root, client, approver, recorder)
	if err != nil {
		return "", recorder, err
	}
	answer, _, err := app.Chat(ctx, nil, question)
	return answer, recorder, err
}

func main() {
	approve := flag.Bool("approve-write", false, "批准本次任务中的写入工具")
	root := flag.String("workspace", envOr("WORKSPACE_DIR", filepath.FromSlash("testdata/workspace")), "工作区目录")
	showTrace := flag.Bool("trace", true, "输出 trace JSON 到 stderr")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: go run ./ch13-capstone [flags] <question>")
		os.Exit(2)
	}
	client := llm.NewHTTPClient(os.Getenv("LLM_API_KEY"), envOr("LLM_BASE_URL", "https://api.deepseek.com"), os.Getenv("LLM_MODEL"))
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	answer, recorder, err := runCapstone(ctx, client, *root, strings.Join(flag.Args(), " "), *approve)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(answer)
	if *showTrace {
		_ = recorder.Export(os.Stderr)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

