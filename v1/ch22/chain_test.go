package main

import (
	"context"
	"testing"
)

type staticChain struct{ out string }

func (s staticChain) Call(_ context.Context, _ string) (string, error) {
	return s.out, nil
}

func TestSequentialChain(t *testing.T) {
	seq := &SequentialChain{Chains: []Chain{staticChain{"A"}, staticChain{"B"}}}
	out, err := seq.Call(context.Background(), "input")
	if err != nil {
		t.Fatal(err)
	}
	if out != "B" {
		t.Fatalf("应串行传递: %q", out)
	}
}

type scriptedLLM struct {
	replies []string
	idx     int
}

func (s *scriptedLLM) Chat(_ context.Context, _ []Message) (Message, error) {
	if s.idx < len(s.replies) {
		r := s.replies[s.idx]
		s.idx++
		return Message{Role: "assistant", Content: r}, nil
	}
	return Message{Role: "assistant", Content: "Final Answer: 完成"}, nil
}

func TestAgentExecutorLoop(t *testing.T) {
	llm := &scriptedLLM{replies: []string{
		"Thought: 需要计算。\nAction: calculator\nAction Input: {\"expression\": \"1 + 2\"}",
		"Thought: 完成。\nFinal Answer: 3",
	}}
	ex := &AgentExecutor{
		LLM: llm,
		Tools: []Tool{{Name: "calculator", Description: "计算", Run: func(_ context.Context, _ string) (string, error) {
			return "3", nil
		}}},
		MaxSteps: 5,
	}
	out, err := ex.Call(context.Background(), "1+2=?")
	if err != nil {
		t.Fatal(err)
	}
	if out != "3" {
		t.Fatalf("最终答案应为 3: %q", out)
	}
}

func TestMemory(t *testing.T) {
	m := &Memory{}
	m.Add("user", "hi")
	m.Add("assistant", "hello")
	if len(m.Messages) != 2 {
		t.Fatalf("记忆条数错误: %d", len(m.Messages))
	}
}
