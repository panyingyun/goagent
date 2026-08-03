package main

import (
	"context"
	"testing"
)

type scripted struct {
	replies []string
	idx     int
}

func (s *scripted) Chat(_ context.Context, _ []Message) (Message, error) {
	if s.idx < len(s.replies) {
		r := s.replies[s.idx]
		s.idx++
		return Message{Role: "assistant", Content: r}, nil
	}
	return Message{Role: "assistant", Content: "通过"}, nil
}

// 一轮通过:规划 -> 执行 -> 审查通过。
func TestOrchestraOneRoundPass(t *testing.T) {
	orch := &Orchestra{
		Planner:   &Agent{Name: "p", Role: "规划", LLM: &scripted{replies: []string{"计划:三步"}}},
		Executor:  &Agent{Name: "e", Role: "执行", LLM: &scripted{replies: []string{"结果内容"}}},
		Reviewer:  &Agent{Name: "r", Role: "审查", LLM: &scripted{replies: []string{"质量合格,通过"}}},
		MaxRounds: 3,
	}
	out, err := orch.Run(context.Background(), "任务")
	if err != nil {
		t.Fatal(err)
	}
	if out != "结果内容" {
		t.Fatalf("最终交付应为执行结果: %q", out)
	}
}

// 审查不通过时,会返工;第二轮通过。
func TestOrchestraReworks(t *testing.T) {
	orch := &Orchestra{
		Planner:   &Agent{Name: "p", Role: "规划", LLM: &scripted{replies: []string{"计划"}}},
		Executor:  &Agent{Name: "e", Role: "执行", LLM: &scripted{replies: []string{"v1", "v2"}}},
		Reviewer:  &Agent{Name: "r", Role: "审查", LLM: &scripted{replies: []string{"不够好,重写", "通过"}}},
		MaxRounds: 3,
	}
	out, err := orch.Run(context.Background(), "任务")
	if err != nil {
		t.Fatal(err)
	}
	if out != "v2" {
		t.Fatalf("返工后应交付 v2: %q", out)
	}
}
