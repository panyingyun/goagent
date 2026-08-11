package structured

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

const (
	MinSteps = 1
	MaxSteps = 5
)

// Step 是任务计划中的单个步骤。
type Step struct {
	ID   string `json:"id"`
	Task string `json:"task"`
}

// TaskPlan 是模型应产出的结构化任务计划(标题 + 有限步骤)。
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
	if len(plan.Steps) < MinSteps || len(plan.Steps) > MaxSteps {
		return plan, fmt.Errorf("steps must contain %d to %d items", MinSteps, MaxSteps)
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
