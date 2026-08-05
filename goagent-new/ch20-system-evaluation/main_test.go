package main

import (
	"context"
	"testing"
)

func TestRunSuitePassesIncludingFaultCase(t *testing.T) {
	report := runSuite(context.Background())
	if report.PassRate != 1 || len(report.Results) != 3 {
		t.Fatalf("unexpected report: %+v", report)
	}
}
