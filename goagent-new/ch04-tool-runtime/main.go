package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"goagent-new/pkg/agent"
	"goagent-new/pkg/llm"
)

func buildExecutor(approve bool, writes *int) (*agent.Executor, error) {
	executor := agent.NewExecutor()
	executor.Approver = agent.ApproverFunc(func(context.Context, agent.ApprovalRequest) (bool, error) {
		return approve, nil
	})
	if err := executor.Register(agent.NewCalculatorTool()); err != nil {
		return nil, err
	}
	writeTool := agent.Tool{
		Name: "save_result", Description: "保存结果的模拟写工具。",
		Parameters: map[string]any{"type": "object", "additionalProperties": false},
		SideEffect: agent.SideEffectWrite, Idempotent: true,
		Run: func(context.Context, json.RawMessage) (string, error) {
			*writes++
			return "saved", nil
		},
	}
	if err := executor.Register(writeTool); err != nil {
		return nil, err
	}
	return executor, nil
}

func main() {
	approve := flag.Bool("approve", false, "批准模拟写操作")
	flag.Parse()
	writes := 0
	executor, err := buildExecutor(*approve, &writes)
	if err != nil {
		panic(err)
	}
	calls := []llm.ToolCall{
		{ID: "read-1", Function: llm.FunctionCall{Name: "calculator", Arguments: `{"expression":"10+5"}`}},
		{ID: "write-1", Function: llm.FunctionCall{Name: "save_result", Arguments: `{}`}},
	}
	for _, result := range executor.ExecuteBatch(context.Background(), calls) {
		fmt.Printf("tool=%s out=%q err=%v\n", result.Call.Function.Name, result.Out, result.Err)
	}
	fmt.Println("实际写入次数:", writes)
}

