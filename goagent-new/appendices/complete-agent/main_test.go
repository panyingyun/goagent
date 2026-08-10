package main

import (
	"context"
	"errors"
	"testing"

	"goagent-new/pkg/agent"
	"goagent-new/pkg/llm"
)

type scriptedClient struct {
	responses []llm.Response
}

func (s *scriptedClient) Complete(_ context.Context, _ llm.Request) (llm.Response, error) {
	if len(s.responses) == 0 {
		return llm.Response{}, errors.New("script exhausted")
	}
	resp := s.responses[0]
	s.responses = s.responses[1:]
	return resp, nil
}

func TestExampleWiresPkgFramework(t *testing.T) {
	ex := agent.NewExecutor()
	if err := ex.Register(agent.NewCalculatorTool()); err != nil {
		t.Fatal(err)
	}
	if err := ex.Register(agent.NewGetTimeTool()); err != nil {
		t.Fatal(err)
	}

	client := &scriptedClient{responses: []llm.Response{
		{Message: llm.Message{Role: llm.RoleAssistant, ToolCalls: []llm.ToolCall{{
			ID: "c1", Type: "function",
			Function: llm.FunctionCall{Name: "calculator", Arguments: `{"expression":"1+1"}`},
		}}}},
		{Message: llm.Message{Role: llm.RoleAssistant, Content: "2"}},
	}}
	runner := agent.Runner{LLM: client, Executor: ex, MaxRounds: 3}
	answer, _, err := runner.Run(context.Background(), []llm.Message{
		{Role: llm.RoleSystem, Content: systemPrompt},
		{Role: llm.RoleUser, Content: "1+1?"},
	})
	if err != nil || answer != "2" {
		t.Fatalf("runner: answer=%q err=%v", answer, err)
	}
}
