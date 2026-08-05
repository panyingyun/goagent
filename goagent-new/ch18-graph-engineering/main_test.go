package main

import (
	"context"
	"testing"

	"goagent-new/pkg/graph"
)

func TestSupportGraphRoutesHighSeverityToHuman(t *testing.T) {
	state, err := runDemo(context.Background(), "high")
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != graph.Succeeded || len(state.Path) != 2 || state.Path[1] != "escalate" {
		t.Fatalf("unexpected state: %+v", state)
	}
}

func TestSupportGraphRoutesNormalRequestToAnswer(t *testing.T) {
	state, err := runDemo(context.Background(), "normal")
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Path) != 2 || state.Path[1] != "answer" {
		t.Fatalf("unexpected path: %v", state.Path)
	}
}
