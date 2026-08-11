package structured

import (
	"context"
	"testing"

	"goagent-new/pkg/llm"
)

type sequenceLLM struct {
	contents []string
	calls    int
	formats  []string
}

func (s *sequenceLLM) Complete(_ context.Context, req llm.Request) (llm.Response, error) {
	format := ""
	if req.ResponseFormat != nil {
		format = req.ResponseFormat.Type
	}
	s.formats = append(s.formats, format)
	if s.calls >= len(s.contents) {
		return llm.Response{}, context.Canceled
	}
	content := s.contents[s.calls]
	s.calls++
	return llm.Response{Message: llm.Message{Role: llm.RoleAssistant, Content: content}}, nil
}

func TestCreateTaskPlanJSONModeAndRepair(t *testing.T) {
	client := &sequenceLLM{contents: []string{
		`528`,
		`{"title":"计算 88*91","steps":[{"id":"s1","task":"拆成 88*(90+1)"},{"id":"s2","task":"分别相乘后相加"}]}`,
	}}
	plan, err := CreateTaskPlan(context.Background(), client, "88*91=?")
	if err != nil {
		t.Fatal(err)
	}
	if plan.Title == "" || len(plan.Steps) != 2 || client.calls != 2 {
		t.Fatalf("plan=%+v calls=%d", plan, client.calls)
	}
	if len(client.formats) == 0 || client.formats[0] != "json_object" {
		t.Fatalf("expected json_object response_format, got %v", client.formats)
	}
}

func TestCreateTaskPlanExhaustsAttempts(t *testing.T) {
	client := &sequenceLLM{contents: []string{`1`, `2`, `3`}}
	_, err := CreateTaskPlan(context.Background(), client, "随便")
	if err == nil {
		t.Fatal("expected failure after max attempts")
	}
	if client.calls != MaxAttempts {
		t.Fatalf("calls=%d want %d", client.calls, MaxAttempts)
	}
}
