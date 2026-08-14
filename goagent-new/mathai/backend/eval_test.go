package main

import (
	"context"
	"testing"

	"mathai/pkg/eval"
)

func TestEvalCasesFromTestdata(t *testing.T) {
	fileCases, err := loadEvalFileCases()
	if err != nil {
		t.Fatal(err)
	}
	cases, err := evalCasesFromFile(fileCases)
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 3 {
		t.Fatalf("loaded %d eval cases, want 3", len(cases))
	}
}

func TestLiveEvalAcceptsEquivalentAddEquations(t *testing.T) {
	cases := []eval.Case{{
		Name:          "trigonometric-viewport",
		Input:         "trig",
		ExpectedTools: []string{"replace_equations", "set_viewport"},
		ExpectedState: map[string]string{"expressions": "cos(x),sin(x)"},
	}}
	results := matchEvalCases(context.Background(), cases, func(context.Context, string) (eval.Outcome, error) {
		return outcomeFromChat(ChatResult{
			Operations: []string{"add_equation", "add_equation", "set_analysis", "set_viewport"},
			State: GraphState{Equations: []Equation{
				{Expression: "sin(x)"},
				{Expression: "cos(x)"},
			}},
		}), nil
	})
	if rate := eval.PassRate(results); rate != 1 {
		t.Fatalf("pass rate = %v, results=%+v", rate, results)
	}
}
