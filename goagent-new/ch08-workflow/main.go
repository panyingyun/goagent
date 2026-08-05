package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"goagent-new/pkg/workflow"
)

type fixedApprover bool

func (a fixedApprover) Approve(context.Context, workflow.State, workflow.Step) (bool, error) {
	return bool(a), nil
}

func newDemoEngine(checkpoint string, approve bool) *workflow.Engine {
	return &workflow.Engine{
		Checkpoint: checkpoint,
		Approver:   fixedApprover(approve),
		Handlers: map[string]workflow.Handler{
			"plan": func(_ context.Context, state *workflow.State, _ *workflow.Step) (string, error) {
				return "planned: " + state.Goal, nil
			},
			"write": func(context.Context, *workflow.State, *workflow.Step) (string, error) {
				return "simulated external write", nil
			},
		},
	}
}

func main() {
	checkpoint := flag.String("checkpoint", filepath.Join("data", "workflow.json"), "检查点路径")
	resume := flag.Bool("resume", false, "从已有检查点恢复")
	approve := flag.Bool("approve", false, "批准有副作用的步骤")
	flag.Parse()
	engine := newDemoEngine(*checkpoint, *approve)
	var state *workflow.State
	var err error
	if *resume {
		state, err = workflow.Load(*checkpoint)
	} else {
		state = workflow.NewState("demo", "生成并保存一份报告", []workflow.Step{
			{ID: "plan"},
			{ID: "write", RequiresApproval: true, IdempotencyKey: "report:v1"},
		})
	}
	if err == nil {
		err = engine.Run(context.Background(), state)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("workflow=%s current=%d steps=%+v\n", state.ID, state.Current, state.Steps)
}

