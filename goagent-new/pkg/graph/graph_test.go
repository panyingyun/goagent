package graph

import (
	"context"
	"strconv"
	"testing"
)

func TestGraphSupportsConditionAndLoop(t *testing.T) {
	definition := Definition{Name: "loop", Version: 1, Entry: "increment", Nodes: map[string]Node{}}
	definition.Nodes["increment"] = Node{ID: "increment", Run: func(_ context.Context, state *State) (string, error) {
		value, _ := strconv.Atoi(state.Data["count"])
		value++
		state.Data["count"] = strconv.Itoa(value)
		return state.Data["count"], nil
	}}
	definition.Nodes["done"] = Node{ID: "done", Run: func(context.Context, *State) (string, error) { return "ok", nil }}
	definition.Edges = []Edge{
		{From: "increment", To: "increment", Name: "again", When: func(state State) bool { return state.Data["count"] != "3" }},
		{From: "increment", To: "done", Name: "finish", When: func(state State) bool { return state.Data["count"] == "3" }},
	}
	state := NewState(definition, "run-1", nil)
	if err := (Engine{Definition: definition, MaxSteps: 6}).Run(context.Background(), state); err != nil {
		t.Fatal(err)
	}
	if state.Status != Succeeded || state.Data["count"] != "3" || len(state.Path) != 4 {
		t.Fatalf("unexpected state: %+v", state)
	}
}

func TestGraphRejectsAmbiguousRoute(t *testing.T) {
	nop := func(context.Context, *State) (string, error) { return "", nil }
	definition := Definition{Name: "ambiguous", Version: 1, Entry: "a", Nodes: map[string]Node{"a": {ID: "a", Run: nop}, "b": {ID: "b", Run: nop}, "c": {ID: "c", Run: nop}}, Edges: []Edge{{From: "a", To: "b", Name: "b"}, {From: "a", To: "c", Name: "c"}}}
	state := NewState(definition, "run-1", nil)
	if err := (Engine{Definition: definition}).Run(context.Background(), state); err == nil {
		t.Fatal("ambiguous graph unexpectedly succeeded")
	}
}

func TestGraphValidationRejectsUnreachableNode(t *testing.T) {
	nop := func(context.Context, *State) (string, error) { return "", nil }
	definition := Definition{Name: "bad", Version: 1, Entry: "a", Nodes: map[string]Node{"a": {ID: "a", Run: nop}, "orphan": {ID: "orphan", Run: nop}}}
	if err := definition.Validate(); err == nil {
		t.Fatal("unreachable node unexpectedly accepted")
	}
}

func TestGraphRejectsWhenNoConditionalRouteMatches(t *testing.T) {
	nop := func(context.Context, *State) (string, error) { return "", nil }
	definition := Definition{Name: "no-route", Version: 1, Entry: "a", Nodes: map[string]Node{"a": {ID: "a", Run: nop}, "b": {ID: "b", Run: nop}}, Edges: []Edge{{From: "a", To: "b", Name: "never", When: func(State) bool { return false }}}}
	state := NewState(definition, "run-1", nil)
	if err := (Engine{Definition: definition}).Run(context.Background(), state); err == nil {
		t.Fatal("missing conditional route unexpectedly completed as a terminal node")
	}
}
