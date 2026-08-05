package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"goagent-new/pkg/harness"
	"goagent-new/pkg/learning"
)

func TestRepoPilotAppliesApprovedPatchAndLearns(t *testing.T) {
	root := createWorkspace(t, "status = TODO\n")
	store, _ := learning.Open("")
	prior, err := store.Propose(learning.Experience{TaskType: "repository_patch", Situation: "previous patch", Lesson: "always verify the resulting file", Scope: root, Confidence: 1, Evidence: []string{"trace-1", "trace-2"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Verify(prior.ID, learning.Gate{MinEvidence: 2, MinConfidence: 0.9, MinPassRate: 1}, learning.Review{Passed: 2, Total: 2}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Promote(prior.ID, false); err != nil {
		t.Fatal(err)
	}
	pilot := &RepoPilot{Root: root, CheckpointDir: filepath.Join(root, ".runs"), Experiences: store}
	result, err := pilot.Run(context.Background(), ChangeRequest{ID: "task-1", File: "sample.txt", Old: "TODO", New: "DONE", Approve: true})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(root, "sample.txt"))
	if string(data) != "status = DONE\n" || result.Harness.Status != harness.Succeeded || result.Graph.Data["verified"] != "yes" {
		t.Fatalf("unexpected result: content=%q result=%+v", data, result)
	}
	if result.Experience == nil || result.Experience.Status != learning.Candidate {
		t.Fatalf("a single run should only create a candidate experience: %+v", result.Experience)
	}
	if result.RetrievedExperiences != 1 {
		t.Fatalf("expected one promoted prior experience, got %d", result.RetrievedExperiences)
	}
}

func TestRepoPilotDoesNotWriteWithoutApproval(t *testing.T) {
	root := createWorkspace(t, "status = TODO\n")
	result, err := (&RepoPilot{Root: root}).Run(context.Background(), ChangeRequest{ID: "task-2", File: "sample.txt", Old: "TODO", New: "DONE", Approve: false})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(root, "sample.txt"))
	if string(data) != "status = TODO\n" || result.Graph.Data["denied"] != "yes" || result.Experience != nil {
		t.Fatalf("unapproved request changed state: content=%q result=%+v", data, result)
	}
}

func TestRepoPilotRejectsPathEscape(t *testing.T) {
	root := createWorkspace(t, "status = TODO\n")
	_, err := (&RepoPilot{Root: root}).Run(context.Background(), ChangeRequest{ID: "task-3", File: filepath.Join("..", "outside.txt"), Old: "TODO", New: "DONE", Approve: true})
	if err == nil || !strings.Contains(err.Error(), "workspace") {
		t.Fatalf("path escape error=%v", err)
	}
}

func createWorkspace(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "sample.txt"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}
