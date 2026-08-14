package eval

import (
	"context"
	"fmt"
	"slices"
)

// ToolUse 记录评估中观察到的一次工具调用。
type ToolUse struct {
	Name      string
	Arguments string
}

// Outcome 是被测 Agent 对单条输入的运行结果。
type Outcome struct {
	Answer string
	Tools  []ToolUse
	State  map[string]string
}

// Case 是一条回归用例(期望工具轨迹与状态断言)。
type Case struct {
	Name          string
	Input         string
	ExpectedTools []string
	ExpectedState map[string]string
}

// RunFunc 执行单条用例并返回 Outcome。
type RunFunc func(context.Context, string) (Outcome, error)

// Result 是单条用例的通过/失败结论。
type Result struct {
	Case   string
	Passed bool
	Reason string
}

func Run(ctx context.Context, cases []Case, run RunFunc) []Result {
	results := make([]Result, 0, len(cases))
	for _, c := range cases {
		out, err := run(ctx, c.Input)
		if err != nil {
			results = append(results, Result{Case: c.Name, Reason: err.Error()})
			continue
		}
		actualTools := make([]string, len(out.Tools))
		for i, use := range out.Tools {
			actualTools[i] = use.Name
		}
		if !slices.Equal(actualTools, c.ExpectedTools) {
			results = append(results, Result{Case: c.Name, Reason: fmt.Sprintf("tool trajectory: got %v want %v", actualTools, c.ExpectedTools)})
			continue
		}
		stateOK := true
		for key, want := range c.ExpectedState {
			if out.State[key] != want {
				results = append(results, Result{Case: c.Name, Reason: fmt.Sprintf("state[%s]=%q want %q", key, out.State[key], want)})
				stateOK = false
				break
			}
		}
		if stateOK {
			results = append(results, Result{Case: c.Name, Passed: true})
		}
	}
	return results
}

func PassRate(results []Result) float64 {
	if len(results) == 0 {
		return 0
	}
	passed := 0
	for _, result := range results {
		if result.Passed {
			passed++
		}
	}
	return float64(passed) / float64(len(results))
}
