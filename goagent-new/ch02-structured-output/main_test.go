package main

import (
	"context"
	"testing"

	"goagent-new/pkg/llm"
)

type staticLLM struct{ content string }

func (s staticLLM) Complete(context.Context, llm.Request) (llm.Response, error) {
	return llm.Response{Message: llm.Message{Role: llm.RoleAssistant, Content: s.content}}, nil
}

func TestCreatePlanValidatesBusinessRules(t *testing.T) {
	plan, err := createPlan(context.Background(), staticLLM{content: `{"title":"发布","steps":[{"id":"s1","task":"测试"}]}`}, "发布应用")
	if err != nil || plan.Title != "发布" {
		t.Fatalf("plan=%+v err=%v", plan, err)
	}
	if _, err := createPlan(context.Background(), staticLLM{content: `{"title":"发布","steps":[]}`}, "发布应用"); err == nil {
		t.Fatal("empty steps should be rejected")
	}
}

