package main

import (
	"context"
	"testing"

	"goagent-new/pkg/llm"
)

type sequenceLLM struct{ responses []llm.Response }

func (s *sequenceLLM) Complete(context.Context, llm.Request) (llm.Response, error) {
	response := s.responses[0]
	s.responses = s.responses[1:]
	return response, nil
}

func TestRunAgentExecutesToolAndReturnsAnswer(t *testing.T) {
	client := &sequenceLLM{responses: []llm.Response{
		{Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{ID: "c1", Type: "function", Function: llm.FunctionCall{Name: "calculator", Arguments: `{"expression":"2+3*4"}`}}}}},
		{Message: llm.Message{Role: llm.RoleAssistant, Content: "14"}},
	}}
	answer, history, err := runAgent(context.Background(), client, "2+3*4")
	if err != nil {
		t.Fatal(err)
	}
	if answer != "14" || len(history) != 5 || history[3].Content != "14" {
		t.Fatalf("answer=%q history=%+v", answer, history)
	}
}
