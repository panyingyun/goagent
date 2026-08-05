package main

import (
	"context"
	"strings"
	"testing"
)

type scripted struct {
	reply  string
	prompt string
}

func (s *scripted) Chat(_ context.Context, messages []Message) (Message, error) {
	s.prompt = messages[len(messages)-1].Content
	return Message{Role: "assistant", Content: s.reply}, nil
}

func TestJudgeScores(t *testing.T) {
	llm := &scripted{reply: `好的,{"score": 8.5, "reason": "回答准确全面"}`}
	j := &Judge{LLM: llm, Criteria: []string{"准确性", "完整性"}}
	score, reason, err := j.Score(context.Background(), "q", "a")
	if err != nil {
		t.Fatal(err)
	}
	if score != 8.5 || reason != "回答准确全面" {
		t.Fatalf("评分解析错误: %v %q", score, reason)
	}
	if !strings.Contains(llm.prompt, "准确性") || !strings.Contains(llm.prompt, "完整性") {
		t.Fatalf("评分标准应出现在提示词里: %q", llm.prompt)
	}
}

func TestJudgeClampsScore(t *testing.T) {
	llm := &scripted{reply: `{"score": 99, "reason": "离谱"}`}
	j := &Judge{LLM: llm}
	score, _, err := j.Score(context.Background(), "q", "a")
	if err != nil {
		t.Fatal(err)
	}
	if score != 10 {
		t.Fatalf("分数应被夹到 10,实际 %v", score)
	}
}

func TestJudgeHandlesNoJSON(t *testing.T) {
	llm := &scripted{reply: "我认为很好"}
	j := &Judge{LLM: llm}
	if _, _, err := j.Score(context.Background(), "q", "a"); err == nil {
		t.Error("没有 JSON 应报错")
	}
}
