package main

import (
	"context"
	"errors"
	"testing"

	"mathai/pkg/llm"
)

type scriptedClient struct {
	responses []llm.Response
	requests  []llm.Request
}

func (c *scriptedClient) Complete(_ context.Context, request llm.Request) (llm.Response, error) {
	c.requests = append(c.requests, request)
	if len(c.responses) == 0 {
		return llm.Response{}, errors.New("no scripted response")
	}
	response := c.responses[0]
	c.responses = c.responses[1:]
	return response, nil
}

func TestMathAgentUsesGraphPatchTool(t *testing.T) {
	t.Parallel()
	client := &scriptedClient{responses: []llm.Response{
		{
			Message: llm.Message{
				Role: llm.RoleAssistant,
				ToolCalls: []llm.ToolCall{{
					ID: "call_1", Type: "function",
					Function: llm.FunctionCall{
						Name:      "apply_graph_patch",
						Arguments: `{"operation":"replace_equations","equations":[{"expression":"x"},{"expression":"x^2"}]}`,
					},
				}},
			},
		},
		{Message: llm.Message{Role: llm.RoleAssistant, Content: "已经画出直线和抛物线。"}},
	}}
	session := Session{ID: "session_test", Version: 1, GraphState: DefaultGraphState()}
	result, err := NewMathAgent(client).Process(context.Background(), session, "比较 y=x 和 y=x²")
	if err != nil {
		t.Fatal(err)
	}
	if result.Answer == "" || len(result.State.Equations) != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.TraceID == "" {
		t.Fatal("trace id is empty")
	}
	if len(client.requests) != 2 || len(client.requests[0].Tools) != 2 {
		t.Fatalf("unexpected requests: %+v", client.requests)
	}
	if len(session.GraphState.Equations) != 0 {
		t.Fatal("agent mutated the input session")
	}
}

func TestGraphPatchToolRejectsUnknownFields(t *testing.T) {
	t.Parallel()
	state := DefaultGraphState()
	_, err := graphPatchTool(&state).Run(context.Background(), []byte(`{"operation":"clear_graph","unexpected":true}`))
	if err == nil {
		t.Fatal("unknown field unexpectedly accepted")
	}
}
