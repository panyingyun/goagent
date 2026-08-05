package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"goagent-new/pkg/memory"
)

func TestExecuteMemoryLifecycle(t *testing.T) {
	store, err := memory.Open(filepath.Join(t.TempDir(), "preferences.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := execute(store, []string{"set", "drink", "coffee"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := execute(store, []string{"get", "drink"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "coffee") || !strings.Contains(out.String(), "source") {
		t.Fatalf("unexpected output: %s", out.String())
	}
	if err := execute(store, []string{"clear"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if len(store.List()) != 0 {
		t.Fatal("clear should remove all preferences")
	}
}

