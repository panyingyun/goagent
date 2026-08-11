package systemeval

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"sync"

	"goagent-new/pkg/graph"
)

// Outcome 是系统评估中一次图运行的观测结果。
type Outcome struct {
	Path       []string          `json:"path"`
	State      map[string]string `json:"state"`
	Steps      int               `json:"steps"`
	Cost       int               `json:"cost"`
	Violations []string          `json:"violations,omitempty"`
	Error      string            `json:"error,omitempty"`
}

// Case 是系统级用例(输入、故障注入与期望路径/状态)。
type Case struct {
	Name          string            `json:"name"`
	Input         map[string]string `json:"input"`
	Faults        map[string]int    `json:"faults,omitempty"`
	ExpectedPath  []string          `json:"expected_path,omitempty"`
	ExpectedState map[string]string `json:"expected_state,omitempty"`
	ExpectError   bool              `json:"expect_error,omitempty"`
	MaxSteps      int               `json:"max_steps,omitempty"`
	MaxCost       int               `json:"max_cost,omitempty"`
}

// RunFunc 执行单条系统评估用例。
type RunFunc func(context.Context, map[string]string, map[string]int) Outcome

// Result 是单条系统评估用例的结论。
type Result struct {
	Case    string  `json:"case"`
	Passed  bool    `json:"passed"`
	Reason  string  `json:"reason,omitempty"`
	Outcome Outcome `json:"outcome"`
}

// Report 汇总一批用例的通过率、步数与成本。
type Report struct {
	Results      []Result `json:"results"`
	PassRate     float64  `json:"pass_rate"`
	AverageSteps float64  `json:"average_steps"`
	TotalCost    int      `json:"total_cost"`
	Violations   int      `json:"violations"`
}

// Comparison 对比基线与候选报告的回归/改进与成本差。
type Comparison struct {
	Regressions  []string `json:"regressions"`
	Improvements []string `json:"improvements"`
	CostDelta    int      `json:"cost_delta"`
	StepDelta    float64  `json:"step_delta"`
}

func Run(ctx context.Context, cases []Case, run RunFunc) Report {
	report := Report{Results: make([]Result, 0, len(cases))}
	passed, steps := 0, 0
	for _, testCase := range cases {
		outcome := run(ctx, cloneMap(testCase.Input), cloneIntMap(testCase.Faults))
		result := Result{Case: testCase.Name, Outcome: outcome}
		result.Reason = evaluate(testCase, outcome)
		result.Passed = result.Reason == ""
		if result.Passed {
			passed++
		}
		steps += outcome.Steps
		report.TotalCost += outcome.Cost
		report.Violations += len(outcome.Violations)
		report.Results = append(report.Results, result)
	}
	if len(cases) > 0 {
		report.PassRate = float64(passed) / float64(len(cases))
		report.AverageSteps = float64(steps) / float64(len(cases))
	}
	return report
}

func Compare(baseline, candidate Report) Comparison {
	comparison := Comparison{Regressions: []string{}, Improvements: []string{}, CostDelta: candidate.TotalCost - baseline.TotalCost, StepDelta: candidate.AverageSteps - baseline.AverageSteps}
	base := make(map[string]bool, len(baseline.Results))
	candidateCases := make(map[string]bool, len(candidate.Results))
	for _, result := range baseline.Results {
		base[result.Case] = result.Passed
	}
	for _, result := range candidate.Results {
		candidateCases[result.Case] = true
		was, exists := base[result.Case]
		if !exists {
			continue
		}
		if was && !result.Passed {
			comparison.Regressions = append(comparison.Regressions, result.Case)
		}
		if !was && result.Passed {
			comparison.Improvements = append(comparison.Improvements, result.Case)
		}
	}
	for _, result := range baseline.Results {
		if result.Passed && !candidateCases[result.Case] {
			comparison.Regressions = append(comparison.Regressions, result.Case)
		}
	}
	sort.Strings(comparison.Regressions)
	sort.Strings(comparison.Improvements)
	return comparison
}

func WithFaults(definition graph.Definition, failures map[string]int) graph.Definition {
	cloned := definition
	cloned.Nodes = make(map[string]graph.Node, len(definition.Nodes))
	injector := &faultInjector{remaining: cloneIntMap(failures)}
	for id, node := range definition.Nodes {
		original := node.Run
		nodeID := id
		node.Run = func(ctx context.Context, state *graph.State) (string, error) {
			if injector.take(nodeID) {
				return "", fmt.Errorf("injected failure at node %q", nodeID)
			}
			return original(ctx, state)
		}
		cloned.Nodes[id] = node
	}
	return cloned
}

func evaluate(testCase Case, outcome Outcome) string {
	if testCase.ExpectError {
		if outcome.Error == "" {
			return "expected an error"
		}
		return ""
	}
	if outcome.Error != "" {
		return outcome.Error
	}
	if len(outcome.Violations) > 0 {
		return fmt.Sprintf("safety violations: %v", outcome.Violations)
	}
	if testCase.ExpectedPath != nil && !slices.Equal(outcome.Path, testCase.ExpectedPath) {
		return fmt.Sprintf("path=%v want %v", outcome.Path, testCase.ExpectedPath)
	}
	for key, want := range testCase.ExpectedState {
		if got := outcome.State[key]; got != want {
			return fmt.Sprintf("state[%s]=%q want %q", key, got, want)
		}
	}
	if testCase.MaxSteps > 0 && outcome.Steps > testCase.MaxSteps {
		return fmt.Sprintf("steps=%d exceed %d", outcome.Steps, testCase.MaxSteps)
	}
	if testCase.MaxCost > 0 && outcome.Cost > testCase.MaxCost {
		return fmt.Sprintf("cost=%d exceeds %d", outcome.Cost, testCase.MaxCost)
	}
	return ""
}

// faultInjector 按节点剩余次数注入故障,用于故障注入评估。
type faultInjector struct {
	mu        sync.Mutex
	remaining map[string]int
}

func (i *faultInjector) take(node string) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.remaining[node] <= 0 {
		return false
	}
	i.remaining[node]--
	return true
}

func cloneMap(input map[string]string) map[string]string {
	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func cloneIntMap(input map[string]int) map[string]int {
	result := make(map[string]int, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}
