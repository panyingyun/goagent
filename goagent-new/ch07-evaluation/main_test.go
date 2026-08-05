package main

import (
	"context"
	"testing"

	"goagent-new/pkg/eval"
)

func TestEvaluationSuite(t *testing.T) {
	results := runSuite(context.Background())
	if rate := eval.PassRate(results); rate != 1 {
		t.Fatalf("pass rate = %v, results=%+v", rate, results)
	}
}

