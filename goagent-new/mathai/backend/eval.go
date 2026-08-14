package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"mathai/pkg/eval"
	"mathai/pkg/llm"
)

type evalFileCase struct {
	ID                    string   `json:"id"`
	Message               string   `json:"message"`
	ExpectedOperations    []string `json:"expectedOperations"`
	ExpectedExpressions   []string `json:"expectedExpressions"`
	ExpectedEquationCount *int     `json:"expectedEquationCount"`
}

func loadEvalFileCases() ([]evalFileCase, error) {
	data, err := os.ReadFile(evalCasesPath())
	if err != nil {
		return nil, err
	}
	var cases []evalFileCase
	if err := json.Unmarshal(data, &cases); err != nil {
		return nil, err
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("no eval cases in %s", evalCasesPath())
	}
	return cases, nil
}

func evalCasesPath() string {
	_, thisFile, _, ok := runtime.Caller(0)
	if ok {
		candidate := filepath.Join(filepath.Dir(thisFile), "..", "testdata", "eval-cases.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	for _, candidate := range []string{
		filepath.Join("..", "testdata", "eval-cases.json"),
		filepath.Join("testdata", "eval-cases.json"),
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return filepath.Join("..", "testdata", "eval-cases.json")
}

func evalCasesFromFile(fileCases []evalFileCase) ([]eval.Case, error) {
	cases := make([]eval.Case, 0, len(fileCases))
	for _, fileCase := range fileCases {
		state := map[string]string{}
		if len(fileCase.ExpectedExpressions) > 0 {
			expressions, err := joinExpectedExpressions(fileCase.ExpectedExpressions)
			if err != nil {
				return nil, err
			}
			state["expressions"] = expressions
		}
		if fileCase.ExpectedEquationCount != nil {
			state["equationCount"] = strconv.Itoa(*fileCase.ExpectedEquationCount)
		}
		cases = append(cases, eval.Case{
			Name:          fileCase.ID,
			Input:         fileCase.Message,
			ExpectedTools: fileCase.ExpectedOperations,
			ExpectedState: state,
		})
	}
	return cases, nil
}

func joinExpectedExpressions(values []string) (string, error) {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		expression, err := NormalizeExpression(value)
		if err != nil {
			return "", err
		}
		normalized = append(normalized, expression)
	}
	sort.Strings(normalized)
	return strings.Join(normalized, ","), nil
}

func outcomeFromChat(result ChatResult) eval.Outcome {
	tools := make([]eval.ToolUse, 0, len(result.Operations))
	for _, operation := range canonicalizeGraphOps(result.Operations) {
		tools = append(tools, eval.ToolUse{Name: operation})
	}
	expressions := make([]string, 0, len(result.State.Equations))
	for _, equation := range result.State.Equations {
		expressions = append(expressions, equation.Expression)
	}
	sort.Strings(expressions)
	return eval.Outcome{
		Answer: result.Answer,
		Tools:  tools,
		State: map[string]string{
			"expressions":   strings.Join(expressions, ","),
			"equationCount": strconv.Itoa(len(result.State.Equations)),
		},
	}
}

func canonicalizeGraphOps(ops []string) []string {
	out := make([]string, 0, len(ops))
	pendingAdds := 0
	flushAdds := func() {
		switch {
		case pendingAdds == 1:
			out = append(out, "add_equation")
		case pendingAdds > 1:
			out = append(out, "replace_equations")
		}
		pendingAdds = 0
	}
	for _, op := range ops {
		if op == "add_equation" {
			pendingAdds++
			continue
		}
		flushAdds()
		out = append(out, op)
	}
	flushAdds()
	return out
}

func isSubsequence(want, got []string) bool {
	if len(want) == 0 {
		return true
	}
	index := 0
	for _, item := range got {
		if item == want[index] {
			index++
			if index == len(want) {
				return true
			}
		}
	}
	return false
}

func matchEvalCases(ctx context.Context, cases []eval.Case, run eval.RunFunc) []eval.Result {
	results := make([]eval.Result, 0, len(cases))
	for _, testCase := range cases {
		out, err := run(ctx, testCase.Input)
		fmt.Println("input = ", testCase.Input)
		if encoded, encodeErr := json.MarshalIndent(out, "", "  "); encodeErr == nil {
			fmt.Println(string(encoded))
		}
		if err != nil {
			results = append(results, eval.Result{Case: testCase.Name, Reason: err.Error()})
			continue
		}
		actual := make([]string, len(out.Tools))
		for i, use := range out.Tools {
			actual[i] = use.Name
		}
		fmt.Println("actual = ", actual)
		if !isSubsequence(testCase.ExpectedTools, actual) {
			results = append(results, eval.Result{Case: testCase.Name, Reason: fmt.Sprintf("tool trajectory: got %v want subsequence %v", actual, testCase.ExpectedTools)})
			continue
		}
		stateOK := true
		for key, want := range testCase.ExpectedState {
			if out.State[key] != want {
				results = append(results, eval.Result{Case: testCase.Name, Reason: fmt.Sprintf("state[%s]=%q want %q", key, out.State[key], want)})
				stateOK = false
				break
			}
		}
		if stateOK {
			results = append(results, eval.Result{Case: testCase.Name, Passed: true})
		}
	}
	return results
}

func runEvalSuite(ctx context.Context, client llm.Client) ([]eval.Result, error) {
	fileCases, err := loadEvalFileCases()
	if err != nil {
		return nil, err
	}
	cases, err := evalCasesFromFile(fileCases)
	if err != nil {
		return nil, err
	}
	run := func(ctx context.Context, input string) (eval.Outcome, error) {
		result, err := NewMathAgent(client).Process(ctx, Session{ID: "eval", Version: 1, GraphState: DefaultGraphState()}, input)
		if err != nil {
			return eval.Outcome{}, err
		}
		return outcomeFromChat(result), nil
	}
	return matchEvalCases(ctx, cases, run), nil
}

func printEvalResults(results []eval.Result) {
	for _, result := range results {
		fmt.Printf("case=%q passed=%v reason=%q\n", result.Case, result.Passed, result.Reason)
	}
	fmt.Printf("pass_rate=%.0f%%\n", eval.PassRate(results)*100)
}

func runEvalCommand(client *llm.HTTPClient) error {
	if !client.Configured() {
		return fmt.Errorf("DEEPSEEK_API_KEY is required for -eval")
	}
	results, err := runEvalSuite(context.Background(), client)
	if err != nil {
		return err
	}
	printEvalResults(results)
	if rate := eval.PassRate(results); rate < 0.9 {
		return fmt.Errorf("pass rate %.0f%% is below 90%%", rate*100)
	}
	return nil
}
