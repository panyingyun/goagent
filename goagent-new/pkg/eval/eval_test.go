package eval

import (
	"context"
	"testing"
)

func TestRunChecksTrajectoryAndState(t *testing.T) {
	cases := []Case{{Name: "save", Input: "save", ExpectedTools: []string{"calculate", "write"}, ExpectedState: map[string]string{"saved": "yes"}}}
	results := Run(context.Background(), cases, func(context.Context, string) (Outcome, error) {
		return Outcome{Tools: []ToolUse{{Name: "calculate"}, {Name: "write"}}, State: map[string]string{"saved": "yes"}}, nil
	})
	if PassRate(results) != 1 {
		t.Fatalf("unexpected results: %+v", results)
	}
}

