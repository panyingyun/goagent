package main

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStorePersistsAndRejectsStaleCommit(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "sessions.json")
	store, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.Create("抛物线")
	if err != nil {
		t.Fatal(err)
	}
	state, err := ApplyGraphPatch(session.GraphState, GraphPatch{Operation: "add_equation", Equation: &EquationInput{Expression: "x^2"}})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	updated, err := store.CommitChat(session.ID, session.Version, Message{ID: "u1", Role: "user", Content: "画抛物线", CreatedAt: now}, Message{ID: "a1", Role: "assistant", Content: "完成", CreatedAt: now}, state)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Version != 2 || len(updated.Messages) != 2 {
		t.Fatalf("unexpected committed session: %+v", updated)
	}
	if _, err := store.CommitChat(session.ID, session.Version, Message{}, Message{}, state); !errors.Is(err, ErrConflict) {
		t.Fatalf("stale commit error = %v", err)
	}

	reopened, err := OpenStore(path)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := reopened.Get(session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Version != 2 || len(loaded.GraphState.Equations) != 1 {
		t.Fatalf("unexpected persisted session: %+v", loaded)
	}
}

func TestCloneSessionJSONUsesEmptyArrays(t *testing.T) {
	t.Parallel()
	store, err := OpenStore(filepath.Join(t.TempDir(), "sessions.json"))
	if err != nil {
		t.Fatal(err)
	}
	session, err := store.Create("空会话")
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(session)
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	if strings.Contains(text, `"messages":null`) || strings.Contains(text, `"equations":null`) {
		t.Fatalf("empty slices must encode as [] not null: %s", text)
	}
}
