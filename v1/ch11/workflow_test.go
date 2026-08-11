package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// mockLLM 实现最小聊天接口。
type mockLLM struct {
	replies []string
	idx     int
}

func (m *mockLLM) Chat(_ context.Context, _ []Message) (Message, error) {
	if m.idx < len(m.replies) {
		r := m.replies[m.idx]
		m.idx++
		return Message{Role: "assistant", Content: r}, nil
	}
	return Message{Role: "assistant", Content: "默认回复"}, nil
}

func TestPlanParsesSteps(t *testing.T) {
	llm := &mockLLM{replies: []string{`[{"desc":"收集资料","question":"收集哪些资料?"},{"desc":"写初稿","question":"初稿怎么写?"}]`}}
	w := &Workflow{ID: "t", Goal: "写周报", Checkpoint: filepath.Join(t.TempDir(), "c.json")}
	if err := w.Plan(context.Background(), llm); err != nil {
		t.Fatal(err)
	}
	if len(w.Steps) != 2 || w.Steps[0].Status != StepPending || w.Steps[1].Status != StepPending {
		t.Fatalf("规划结果异常: %+v", w.Steps)
	}
}

func TestCheckpointResume(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c.json")
	w := &Workflow{ID: "t", Goal: "测试", Checkpoint: path,
		Steps: []Step{
			{ID: 0, Desc: "s1", Question: "q1", Status: StepDone, Output: "o1"},
			{ID: 1, Desc: "s2", Question: "q2", Status: StepPending},
			{ID: 2, Desc: "s3", Question: "q3", Status: StepPending},
		},
		Current: 1,
	}
	if err := w.Save(); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadWorkflow(path)
	if err != nil {
		t.Fatal(err)
	}
	next := loaded.NextStep()
	if next == nil || next.ID != 1 {
		t.Fatalf("应从步骤 1 继续,实际 %+v", next)
	}
	llm := &mockLLM{replies: []string{"答案2", "答案3"}}
	if err := loaded.Resume(context.Background(), llm); err != nil {
		t.Fatal(err)
	}
	if loaded.Steps[2].Output != "答案3" || loaded.Current != 3 {
		t.Fatalf("恢复执行异常: %+v", loaded.Steps)
	}
	_ = os.Remove(path)
}
