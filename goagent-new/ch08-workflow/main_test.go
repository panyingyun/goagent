package main

import (
	"context"
	"path/filepath"
	"testing"

	"goagent-new/pkg/workflow"
)

func TestDemoWorkflowWaitsForApprovalThenCompletes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "workflow.json")
	state := workflow.NewState("demo", "goal", []workflow.Step{{ID: "plan"}, {ID: "write", RequiresApproval: true, IdempotencyKey: "write:v1"}})
	if err := newDemoEngine(path, false).Run(context.Background(), state); err == nil {
		t.Fatal("unapproved workflow should stop")
	}
	loaded, err := workflow.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := newDemoEngine(path, true).Run(context.Background(), loaded); err != nil {
		t.Fatal(err)
	}
	if loaded.Current != 2 || loaded.Steps[1].Status != workflow.Done {
		t.Fatalf("workflow did not complete: %+v", loaded)
	}
}

