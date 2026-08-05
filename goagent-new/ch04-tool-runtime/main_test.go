package main

import (
	"context"
	"testing"

	"goagent-new/pkg/llm"
)

func TestWriteRequiresApprovalAndDeduplicates(t *testing.T) {
	writes := 0
	executor, err := buildExecutor(false, &writes)
	if err != nil {
		t.Fatal(err)
	}
	call := llm.ToolCall{ID: "w1", Function: llm.FunctionCall{Name: "save_result", Arguments: `{}`}}
	if result := executor.Execute(context.Background(), call); result.Err == nil || writes != 0 {
		t.Fatalf("unapproved write: %+v writes=%d", result, writes)
	}
	executor, err = buildExecutor(true, &writes)
	if err != nil {
		t.Fatal(err)
	}
	if result := executor.Execute(context.Background(), call); result.Err != nil {
		t.Fatal(result.Err)
	}
	if result := executor.Execute(context.Background(), call); result.Err != nil {
		t.Fatal(result.Err)
	}
	if writes != 1 {
		t.Fatalf("idempotent write executed %d times", writes)
	}
}

