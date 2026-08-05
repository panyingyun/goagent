package systemeval

import (
	"context"
	"testing"
)

func TestRunMeasuresPathStateAndBudgets(t *testing.T) {
	cases := []Case{{Name: "success", Input: map[string]string{"severity": "high"}, ExpectedPath: []string{"classify", "escalate"}, ExpectedState: map[string]string{"route": "human"}, MaxSteps: 2, MaxCost: 3}}
	report := Run(context.Background(), cases, func(_ context.Context, _ map[string]string, _ map[string]int) Outcome {
		return Outcome{Path: []string{"classify", "escalate"}, State: map[string]string{"route": "human"}, Steps: 2, Cost: 2}
	})
	if report.PassRate != 1 || report.AverageSteps != 2 {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestCompareDetectsRegression(t *testing.T) {
	baseline := Report{Results: []Result{{Case: "a", Passed: true}, {Case: "b", Passed: false}, {Case: "missing", Passed: true}}, TotalCost: 5, AverageSteps: 2}
	candidate := Report{Results: []Result{{Case: "a", Passed: false}, {Case: "b", Passed: true}}, TotalCost: 7, AverageSteps: 3}
	comparison := Compare(baseline, candidate)
	if len(comparison.Regressions) != 2 || comparison.Regressions[0] != "a" || comparison.Regressions[1] != "missing" || len(comparison.Improvements) != 1 || comparison.CostDelta != 2 {
		t.Fatalf("unexpected comparison: %+v", comparison)
	}
}
