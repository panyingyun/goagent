package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"goagent-new/pkg/eval"
)

func demoRun(_ context.Context, input string) (eval.Outcome, error) {
	switch {
	case strings.Contains(input, "计算"):
		return eval.Outcome{Answer: "14", Tools: []eval.ToolUse{{Name: "calculator"}}, State: map[string]string{"result": "14"}}, nil
	case strings.Contains(input, "列出"):
		return eval.Outcome{Answer: "handbook.txt", Tools: []eval.ToolUse{{Name: "list_documents"}}, State: map[string]string{"listed": "yes"}}, nil
	default:
		return eval.Outcome{Answer: "直接回答", State: map[string]string{}}, nil
	}
}

func runSuite(ctx context.Context) []eval.Result {
	cases := []eval.Case{
		{Name: "calculator trajectory", Input: "计算 2+3*4", ExpectedTools: []string{"calculator"}, ExpectedState: map[string]string{"result": "14"}},
		{Name: "list trajectory", Input: "列出文档", ExpectedTools: []string{"list_documents"}, ExpectedState: map[string]string{"listed": "yes"}},
		{Name: "no unnecessary tool", Input: "你好", ExpectedTools: nil},
	}
	return eval.Run(ctx, cases, demoRun)
}

func main() {
	results := runSuite(context.Background())
	for _, result := range results {
		fmt.Printf("case=%q passed=%v reason=%q\n", result.Case, result.Passed, result.Reason)
	}
	rate := eval.PassRate(results)
	fmt.Printf("pass_rate=%.0f%%\n", rate*100)
	if rate != 1 {
		os.Exit(1)
	}
}

