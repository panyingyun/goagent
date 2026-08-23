package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"mathai/pkg/llm"
)

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
	if len(client.requests) != 2 || len(client.requests[0].Tools) != 3 {
		t.Fatalf("unexpected requests: %+v", client.requests)
	}
	if !hasToolName(client.requests[0].Tools, "dollar_power") {
		t.Fatal("dollar_power tool was not registered")
	}
	if len(session.GraphState.Equations) != 0 {
		t.Fatal("agent mutated the input session")
	}
}

func TestGraphPatchToolRejectsUnknownFields(t *testing.T) {
	t.Parallel()
	state := DefaultGraphState()
	_, err := graphPatchTool(&state, nil).Run(context.Background(), []byte(`{"operation":"clear_graph","unexpected":true}`))
	if err == nil {
		t.Fatal("unknown field unexpectedly accepted")
	}
}

func TestDollarPowerToolComputesExponentiation(t *testing.T) {
	t.Parallel()
	cases := []struct {
		base, exponent, want float64
	}{
		{1, 5, 1},
		{2, 3, 8},
		{1, 6, 1},
	}
	tool := dollarPowerTool()
	for _, tc := range cases {
		raw, err := tool.Run(context.Background(), []byte(fmt.Sprintf(`{"base":%v,"exponent":%v}`, tc.base, tc.exponent)))
		if err != nil {
			t.Fatalf("%v $ %v: %v", tc.base, tc.exponent, err)
		}
		var out struct {
			Result float64 `json:"result"`
		}
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			t.Fatal(err)
		}
		if out.Result != tc.want {
			t.Fatalf("%v $ %v = %v, want %v", tc.base, tc.exponent, out.Result, tc.want)
		}
	}
}

func TestMathAgentUsesDollarPowerTool(t *testing.T) {
	t.Parallel()
	client := &scriptedClient{responses: []llm.Response{
		{
			Message: llm.Message{
				Role: llm.RoleAssistant,
				ToolCalls: []llm.ToolCall{{
					ID: "call_1", Type: "function",
					Function: llm.FunctionCall{Name: "dollar_power", Arguments: `{"base":2,"exponent":3}`},
				}},
			},
		},
		{Message: llm.Message{Role: llm.RoleAssistant, Content: "`2 $ 3` 表示 $2^{3}$，答案是 **8**。"}},
	}}
	session := Session{ID: "session_dollar", Version: 1, GraphState: DefaultGraphState()}
	result, err := NewMathAgent(client).Process(context.Background(), session, "2$3 等于多少")
	if err != nil {
		t.Fatal(err)
	}
	if result.Answer == "" || !strings.Contains(result.Answer, "8") {
		t.Fatalf("unexpected answer: %q", result.Answer)
	}
	if len(result.State.Equations) != 0 || len(result.Operations) != 0 {
		t.Fatalf("dollar_power must not change the graph: %+v", result)
	}
	if len(client.requests) != 2 {
		t.Fatalf("LLM should be called, got %d requests", len(client.requests))
	}
	if client.requests[0].Messages[len(client.requests[0].Messages)-1].Content != "2$3 等于多少" {
		t.Fatalf("user text was not sent to LLM: %+v", client.requests[0].Messages)
	}
	toolMsg := client.requests[1].Messages[len(client.requests[1].Messages)-1]
	if toolMsg.Role != llm.RoleTool || !strings.Contains(toolMsg.Content, `"result":8`) {
		t.Fatalf("tool result was not returned to LLM: %+v", toolMsg)
	}
}

func hasToolName(tools []llm.ToolDefinition, name string) bool {
	for _, tool := range tools {
		if tool.Function.Name == name {
			return true
		}
	}
	return false
}
