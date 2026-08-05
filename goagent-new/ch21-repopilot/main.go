package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	root := flag.String("workspace", filepath.Join("ch21-repopilot", "testdata", "workspace"), "controlled workspace root")
	file := flag.String("file", "sample.txt", "relative text file")
	oldText := flag.String("old", "TODO", "exact text to replace")
	newText := flag.String("new", "DONE", "replacement text")
	approve := flag.Bool("approve", false, "approve the workspace write")
	flag.Parse()
	pilot := &RepoPilot{Root: *root, CheckpointDir: filepath.Join("data", "repopilot")}
	result, err := pilot.Run(context.Background(), ChangeRequest{ID: "cli-demo", File: *file, Old: *oldText, New: *newText, Approve: *approve})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("status=%s path=%v approved=%v\n", result.Harness.Status, result.Graph.Path, *approve)
	if result.Experience != nil {
		fmt.Printf("learned=%s status=%s\n", result.Experience.ID, result.Experience.Status)
	}
}
