package main

import (
	"context"
	"fmt"

	"goagent-new/pkg/graph"
)

func supportGraph() graph.Definition {
	nop := func(_ context.Context, state *graph.State) (string, error) { return state.Current, nil }
	return graph.Definition{
		Name: "support-router", Version: 1, Entry: "classify",
		Nodes: map[string]graph.Node{
			"classify": {ID: "classify", Run: func(_ context.Context, state *graph.State) (string, error) {
				if state.Data["severity"] == "high" {
					state.Data["route"] = "human"
				} else {
					state.Data["route"] = "answer"
				}
				return state.Data["route"], nil
			}},
			"answer":   {ID: "answer", Run: nop},
			"escalate": {ID: "escalate", Run: nop},
		},
		Edges: []graph.Edge{
			{From: "classify", To: "answer", Name: "normal", When: func(state graph.State) bool { return state.Data["route"] == "answer" }},
			{From: "classify", To: "escalate", Name: "high-risk", When: func(state graph.State) bool { return state.Data["route"] == "human" }},
		},
	}
}

func runDemo(ctx context.Context, severity string) (*graph.State, error) {
	definition := supportGraph()
	state := graph.NewState(definition, "demo", map[string]string{"severity": severity})
	err := (graph.Engine{Definition: definition, MaxSteps: 5}).Run(ctx, state)
	return state, err
}

func main() {
	state, err := runDemo(context.Background(), "high")
	if err != nil {
		panic(err)
	}
	fmt.Printf("status=%s path=%v route=%s\n", state.Status, state.Path, state.Data["route"])
}
