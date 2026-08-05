package structured

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Step struct {
	ID   string `json:"id"`
	Task string `json:"task"`
}

type TaskPlan struct {
	Title string `json:"title"`
	Steps []Step `json:"steps"`
}

func DecodeTaskPlan(content string) (TaskPlan, error) {
	var plan TaskPlan
	decoder := json.NewDecoder(strings.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&plan); err != nil {
		return plan, fmt.Errorf("decode task plan: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return plan, errors.New("task plan must contain exactly one JSON value")
	}
	if strings.TrimSpace(plan.Title) == "" {
		return plan, errors.New("title is required")
	}
	if len(plan.Steps) < 1 || len(plan.Steps) > 5 {
		return plan, errors.New("steps must contain 1 to 5 items")
	}
	seen := make(map[string]bool)
	for i, step := range plan.Steps {
		if strings.TrimSpace(step.ID) == "" || strings.TrimSpace(step.Task) == "" {
			return plan, fmt.Errorf("step %d requires id and task", i)
		}
		if seen[step.ID] {
			return plan, fmt.Errorf("duplicate step id %q", step.ID)
		}
		seen[step.ID] = true
	}
	return plan, nil
}

