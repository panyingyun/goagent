package main

import (
	"context"
	"sync"
	"testing"

	"goagent-new/pkg/llm"
)

type scriptedClient struct {
	mu        sync.Mutex
	responses []llm.Response
}

func (s *scriptedClient) Complete(context.Context, llm.Request) (llm.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	response := s.responses[0]
	s.responses = s.responses[1:]
	return response, nil
}

func TestCapstoneUsesSharedAgentRuntimeAndTrace(t *testing.T) {
	client := &scriptedClient{responses: []llm.Response{
		{Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "c1", Type: "function", Function: llm.FunctionCall{Name: "calculator", Arguments: `{"expression":"12*(8+4)"}`}}}}},
		{Message: llm.Message{Role: llm.RoleAssistant, Content: "144"}},
	}}
	answer, recorder, err := runCapstone(context.Background(), client, t.TempDir(), "计算 12*(8+4)", false)
	if err != nil {
		t.Fatal(err)
	}
	_, events := recorder.Snapshot()
	if answer != "144" || len(events) < 3 {
		t.Fatalf("answer=%q events=%+v", answer, events)
	}
}

