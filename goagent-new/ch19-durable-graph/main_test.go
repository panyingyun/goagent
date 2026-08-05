package main

import (
	"context"
	"testing"

	"goagent-new/pkg/graph"
)

func TestRunDemoPersistsAndResumes(t *testing.T) {
	directory := t.TempDir()
	state, events, err := runDemo(context.Background(), directory, false)
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != graph.Succeeded || len(events) != 4 {
		t.Fatalf("unexpected first run: state=%+v events=%d", state, len(events))
	}
	resumed, afterResume, err := runDemo(context.Background(), directory, true)
	if err != nil {
		t.Fatal(err)
	}
	if resumed.Step != state.Step || len(afterResume) != len(events) {
		t.Fatal("resuming a completed graph should not execute more nodes")
	}
}
