package main

import (
	"context"
	"fmt"

	"goagent-new/pkg/graph"
	"goagent-new/pkg/systemeval"
)

func evaluationGraph() graph.Definition {
	return graph.Definition{
		Name: "risk-router", Version: 1, Entry: "classify",
		Nodes: map[string]graph.Node{
			"classify": {ID: "classify", Run: func(_ context.Context, state *graph.State) (string, error) {
				if state.Data["severity"] == "high" {
					state.Data["route"] = "human"
				} else {
					state.Data["route"] = "auto"
				}
				return state.Data["route"], nil
			}},
			"answer":   {ID: "answer", Run: func(context.Context, *graph.State) (string, error) { return "answered", nil }},
			"escalate": {ID: "escalate", Run: func(context.Context, *graph.State) (string, error) { return "escalated", nil }},
		},
		Edges: []graph.Edge{
			{From: "classify", To: "answer", Name: "auto", When: func(state graph.State) bool { return state.Data["route"] == "auto" }},
			{From: "classify", To: "escalate", Name: "human", When: func(state graph.State) bool { return state.Data["route"] == "human" }},
		},
	}
}

func runSuite(ctx context.Context) systemeval.Report {
	cases := []systemeval.Case{
		{Name: "normal route", Input: map[string]string{"severity": "normal"}, ExpectedPath: []string{"classify", "answer"}, ExpectedState: map[string]string{"route": "auto"}, MaxSteps: 2, MaxCost: 2},
		{Name: "high risk route", Input: map[string]string{"severity": "high"}, ExpectedPath: []string{"classify", "escalate"}, ExpectedState: map[string]string{"route": "human"}, MaxSteps: 2, MaxCost: 2},
		{Name: "injected classifier failure", Input: map[string]string{"severity": "high"}, Faults: map[string]int{"classify": 1}, ExpectError: true},
	}
	return systemeval.Run(ctx, cases, func(ctx context.Context, input map[string]string, faults map[string]int) systemeval.Outcome {
		definition := systemeval.WithFaults(evaluationGraph(), faults)
		state := graph.NewState(definition, "eval", input)
		err := (graph.Engine{Definition: definition, MaxSteps: 4}).Run(ctx, state)
		outcome := systemeval.Outcome{Path: state.Path, State: state.Data, Steps: state.Step, Cost: state.Step}
		if err != nil {
			outcome.Error = err.Error()
		}
		return outcome
	})
}

func main() {
	report := runSuite(context.Background())
	for _, result := range report.Results {
		fmt.Printf("case=%q passed=%v reason=%q path=%v\n", result.Case, result.Passed, result.Reason, result.Outcome.Path)
	}
	fmt.Printf("pass_rate=%.0f%% avg_steps=%.2f total_cost=%d violations=%d\n", report.PassRate*100, report.AverageSteps, report.TotalCost, report.Violations)
}
