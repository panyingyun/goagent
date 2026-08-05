package harness

import (
	"context"
	"errors"
	"testing"
)

func TestRunnerCompletesAndVerifies(t *testing.T) {
	worker := WorkerFunc(func(_ context.Context, _ Task, state *State, _ Environment) (Action, error) {
		state.Values["artifact"] = "report.md"
		return Action{Name: "deliver", Summary: "created report", Progress: true, Done: true, Cost: 2}, nil
	})
	runner := Runner{
		Environment: Environment{Manifest: Manifest{Capabilities: []string{"workspace.read", "workspace.write"}}},
		Worker:      worker,
		Verifiers: []Verifier{VerifierFunc(func(_ context.Context, _ Task, state State) error {
			if state.Values["artifact"] == "" {
				return errors.New("artifact missing")
			}
			return nil
		})},
	}
	state, err := runner.Run(context.Background(), Task{ID: "t1", Goal: "write report", RequiredCapabilities: []string{"workspace.write"}, Budget: Budget{MaxSteps: 3, MaxFailures: 1, MaxCost: 5}})
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != Succeeded || state.Step != 1 || state.Cost != 2 {
		t.Fatalf("unexpected state: %+v", state)
	}
}

func TestRunnerRejectsMissingCapability(t *testing.T) {
	runner := Runner{Worker: WorkerFunc(func(context.Context, Task, *State, Environment) (Action, error) { return Action{}, nil })}
	_, err := runner.Run(context.Background(), Task{ID: "t1", Goal: "write", RequiredCapabilities: []string{"workspace.write"}, Budget: Budget{MaxSteps: 1, MaxFailures: 1}})
	if err == nil {
		t.Fatal("missing capability unexpectedly accepted")
	}
}

func TestRunnerStopsWhenWorkerStalls(t *testing.T) {
	runner := Runner{Worker: WorkerFunc(func(context.Context, Task, *State, Environment) (Action, error) {
		return Action{Name: "thinking"}, nil
	})}
	state, err := runner.Run(context.Background(), Task{ID: "t1", Goal: "finish", Budget: Budget{MaxSteps: 5, MaxFailures: 1}})
	if err == nil || state.Status != Failed || state.Step != 3 {
		t.Fatalf("unexpected stalled result: state=%+v err=%v", state, err)
	}
}
