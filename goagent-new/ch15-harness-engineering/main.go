package main

import (
	"context"
	"errors"
	"fmt"

	"goagent-new/pkg/harness"
)

func runDemo(ctx context.Context) (harness.State, error) {
	stages := []string{"inspect", "plan", "deliver"}
	worker := harness.WorkerFunc(func(_ context.Context, _ harness.Task, state *harness.State, _ harness.Environment) (harness.Action, error) {
		stage := stages[state.Step]
		if stage == "deliver" {
			state.Values["artifact"] = "report.md"
		}
		return harness.Action{Name: stage, Summary: "completed " + stage, Progress: true, Done: stage == "deliver", Cost: 1}, nil
	})
	runner := harness.Runner{
		Environment: harness.Environment{Manifest: harness.Manifest{Name: "course-harness", Version: "1", Capabilities: []string{"workspace.read", "artifact.write"}}},
		Worker:      worker,
		Verifiers: []harness.Verifier{harness.VerifierFunc(func(_ context.Context, _ harness.Task, state harness.State) error {
			if state.Values["artifact"] == "" {
				return errors.New("artifact was not produced")
			}
			return nil
		})},
	}
	return runner.Run(ctx, harness.Task{ID: "demo", Goal: "produce a verified report", RequiredCapabilities: []string{"workspace.read", "artifact.write"}, Budget: harness.Budget{MaxSteps: 5, MaxFailures: 2, MaxCost: 10}})
}

func main() {
	state, err := runDemo(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("status=%s steps=%d cost=%d artifact=%s\n", state.Status, state.Step, state.Cost, state.Values["artifact"])
}
