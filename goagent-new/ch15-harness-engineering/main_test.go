package main

import (
	"context"
	"testing"

	"goagent-new/pkg/harness"
)

func TestRunDemo(t *testing.T) {
	state, err := runDemo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != harness.Succeeded || state.Values["artifact"] != "report.md" {
		t.Fatalf("unexpected state: %+v", state)
	}
}
