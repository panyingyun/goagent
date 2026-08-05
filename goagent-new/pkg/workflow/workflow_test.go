package workflow

import (
	"context"
	"path/filepath"
	"testing"
)

type allow struct{}

func (allow) Approve(context.Context, State, Step) (bool, error) { return true, nil }

func TestWorkflowPersistsAndResumes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	state := NewState("wf-1", "demo", []Step{
		{ID: "read"},
		{ID: "write", RequiresApproval: true, IdempotencyKey: "write:result"},
	})
	writeCount := 0
	engine := Engine{Checkpoint: path, Approver: allow{}, Handlers: map[string]Handler{
		"read": func(context.Context, *State, *Step) (string, error) { return "data", nil },
		"write": func(context.Context, *State, *Step) (string, error) { writeCount++; return "saved", nil },
	}}
	if err := engine.Run(context.Background(), state); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	loaded.Current = 1 // 模拟恢复位置落后，但幂等键已记录。
	loaded.Steps[1].Status = Pending
	if err := engine.Run(context.Background(), loaded); err != nil {
		t.Fatal(err)
	}
	if writeCount != 1 {
		t.Fatalf("side effect repeated %d times", writeCount)
	}
}

