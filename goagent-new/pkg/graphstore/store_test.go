package graphstore

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"goagent-new/pkg/graph"
)

func TestFailedGraphResumesWithExternalIdempotency(t *testing.T) {
	directory := t.TempDir()
	store := &FileStore{CheckpointPath: filepath.Join(directory, "state.json"), EventPath: filepath.Join(directory, "events.jsonl")}
	effects, attempts := 0, 0
	externalResults := make(map[string]string)
	definition := graph.Definition{Name: "durable", Version: 1, Entry: "publish", Nodes: map[string]graph.Node{}}
	definition.Nodes["publish"] = graph.Node{ID: "publish", IdempotencyKey: func(graph.State) string { return "publish:report-v1" }, Run: func(_ context.Context, _ *graph.State) (string, error) {
		attempts++
		key := "publish:report-v1"
		result, exists := externalResults[key]
		if !exists {
			effects++
			result = "remote-42"
			externalResults[key] = result
		}
		if attempts == 1 {
			return "", errors.New("crashed after external success")
		}
		return result, nil
	}}
	engine := graph.Engine{Definition: definition, MaxSteps: 3, Checkpointer: store, Events: store}
	state := graph.NewState(definition, "run-1", nil)
	if err := engine.Run(context.Background(), state); err == nil {
		t.Fatal("first run should fail")
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Run(context.Background(), loaded); err != nil {
		t.Fatal(err)
	}
	if effects != 1 || attempts != 2 || loaded.Status != graph.Succeeded {
		t.Fatalf("unexpected recovery: effects=%d attempts=%d state=%+v", effects, attempts, loaded)
	}
	events, err := store.ReadEvents()
	if err != nil || len(events) < 4 {
		t.Fatalf("events were not persisted: count=%d err=%v", len(events), err)
	}
}

func TestMigrateRequiresEveryVersionStep(t *testing.T) {
	state := &graph.State{GraphVersion: 1, Data: map[string]string{}}
	err := Migrate(state, 3, map[int]Migration{1: func(state *graph.State) error { state.Data["v2"] = "yes"; return nil }})
	if err == nil || state.GraphVersion != 2 {
		t.Fatalf("missing migration unexpectedly succeeded: version=%d err=%v", state.GraphVersion, err)
	}
	if err := Migrate(state, 3, map[int]Migration{2: func(state *graph.State) error { state.Data["v3"] = "yes"; return nil }}); err != nil {
		t.Fatal(err)
	}
	if state.GraphVersion != 3 || state.Data["v3"] != "yes" {
		t.Fatalf("unexpected migrated state: %+v", state)
	}
}
