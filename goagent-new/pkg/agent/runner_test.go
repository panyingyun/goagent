package agent

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"goagent-new/pkg/llm"
)

type scriptedClient struct {
	mu        sync.Mutex
	responses []llm.Response
	requests  []llm.Request
}

func (s *scriptedClient) Complete(_ context.Context, req llm.Request) (llm.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests = append(s.requests, req)
	if len(s.responses) == 0 {
		return llm.Response{}, errors.New("script exhausted")
	}
	resp := s.responses[0]
	s.responses = s.responses[1:]
	return resp, nil
}

func TestRunnerCompletesToolLoop(t *testing.T) {
	executor := NewExecutor()
	if err := executor.Register(NewCalculatorTool()); err != nil {
		t.Fatal(err)
	}
	client := &scriptedClient{responses: []llm.Response{
		{Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "call-1", Type: "function", Function: llm.FunctionCall{Name: "calculator", Arguments: `{"expression":"12*(8+4)"}`}}}}},
		{Message: llm.Message{Role: llm.RoleAssistant, Content: "结果是 144。"}},
	}}
	runner := Runner{LLM: client, Executor: executor, MaxRounds: 3}
	answer, history, err := runner.Run(context.Background(), []llm.Message{{Role: llm.RoleUser, Content: "计算"}})
	if err != nil {
		t.Fatal(err)
	}
	if answer != "结果是 144。" {
		t.Fatalf("answer = %q", answer)
	}
	if len(history) != 4 || history[2].Role != llm.RoleTool || history[2].ToolCallID != "call-1" || history[2].Content != "144" {
		t.Fatalf("unexpected history: %+v", history)
	}
}

func TestWriteToolRequiresApprovalAndIsIdempotent(t *testing.T) {
	executor := NewExecutor()
	writes := 0
	if err := executor.Register(Tool{
		Name: "write_note", Parameters: objectSchema(map[string]any{"content": map[string]any{"type": "string"}}, "content"),
		SideEffect: SideEffectWrite, Idempotent: true,
		Run: func(context.Context, json.RawMessage) (string, error) { writes++; return "saved", nil },
	}); err != nil {
		t.Fatal(err)
	}
	call := llm.ToolCall{ID: "1", Function: llm.FunctionCall{Name: "write_note", Arguments: `{"content":"hello"}`}}
	if result := executor.Execute(context.Background(), call); result.Err == nil {
		t.Fatal("write should be denied without an approver")
	}
	approvals := 0
	executor.Approver = ApproverFunc(func(_ context.Context, _ ApprovalRequest) (bool, error) {
		approvals++
		return true, nil
	})
	if result := executor.Execute(context.Background(), call); result.Err != nil {
		t.Fatal(result.Err)
	}
	if result := executor.Execute(context.Background(), call); result.Err != nil {
		t.Fatal(result.Err)
	}
	if approvals != 1 {
		t.Fatalf("approval count = %d, want 1 because the second call is idempotent", approvals)
	}
	if writes != 1 {
		t.Fatalf("write count = %d, want 1", writes)
	}
}

func TestReadToolsCanRunInParallel(t *testing.T) {
	executor := NewExecutor()
	for _, name := range []string{"one", "two"} {
		name := name
		err := executor.Register(Tool{
			Name: name, Parameters: objectSchema(nil), SideEffect: SideEffectRead, ParallelSafe: true,
			Run: func(ctx context.Context, _ json.RawMessage) (string, error) {
				select {
				case <-time.After(60 * time.Millisecond):
					return name, nil
				case <-ctx.Done():
					return "", ctx.Err()
				}
			},
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	started := time.Now()
	results := executor.ExecuteBatch(context.Background(), []llm.ToolCall{
		{ID: "1", Function: llm.FunctionCall{Name: "one", Arguments: `{}`}},
		{ID: "2", Function: llm.FunctionCall{Name: "two", Arguments: `{}`}},
	})
	if elapsed := time.Since(started); elapsed > 110*time.Millisecond {
		t.Fatalf("tools did not run in parallel: %v", elapsed)
	}
	if results[0].Out != "one" || results[1].Out != "two" {
		t.Fatalf("result order changed: %+v", results)
	}
}

func TestSafePathRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	tool := NewReadFileTool(root)
	_, err := tool.Run(context.Background(), json.RawMessage(`{"path":"../secret.txt"}`))
	if err == nil {
		t.Fatal("expected traversal rejection")
	}
}
