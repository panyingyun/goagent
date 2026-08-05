package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"goagent-new/pkg/graph"
	"goagent-new/pkg/graphstore"
)

func publicationGraph() graph.Definition {
	return graph.Definition{
		Name: "publication", Version: 1, Entry: "prepare",
		Nodes: map[string]graph.Node{
			"prepare": {ID: "prepare", Run: func(_ context.Context, state *graph.State) (string, error) {
				state.Data["draft"] = "ready"
				return "draft ready", nil
			}},
			"publish": {ID: "publish", IdempotencyKey: func(state graph.State) string { return "publish:" + state.RunID }, Run: func(_ context.Context, state *graph.State) (string, error) {
				state.Data["publication_id"] = "pub-42"
				return "pub-42", nil
			}},
		},
		Edges: []graph.Edge{{From: "prepare", To: "publish", Name: "prepared"}},
	}
}

func runDemo(ctx context.Context, directory string, resume bool) (*graph.State, []graph.Event, error) {
	definition := publicationGraph()
	store := &graphstore.FileStore{CheckpointPath: filepath.Join(directory, "publication.json"), EventPath: filepath.Join(directory, "publication.jsonl")}
	var state *graph.State
	var err error
	if resume {
		state, err = store.Load()
	} else {
		state = graph.NewState(definition, "publication-demo", nil)
	}
	if err != nil {
		return nil, nil, err
	}
	engine := graph.Engine{Definition: definition, MaxSteps: 5, Checkpointer: store, Events: store}
	if err := engine.Run(ctx, state); err != nil {
		return state, nil, err
	}
	events, err := store.ReadEvents()
	return state, events, err
}

func main() {
	directory := flag.String("data", filepath.Join("data", "durable-graph"), "checkpoint and event directory")
	resume := flag.Bool("resume", false, "resume the existing checkpoint")
	flag.Parse()
	state, events, err := runDemo(context.Background(), *directory, *resume)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("status=%s path=%v events=%d publication=%s\n", state.Status, state.Path, len(events), state.Data["publication_id"])
}
