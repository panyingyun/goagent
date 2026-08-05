package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
)

// EvalCase 是一个评估用例:问一个问题,检查回答是否包含期望片段。
type EvalCase struct {
	Question    string `json:"question"`
	MustContain string `json:"must_contain"`
}

func loadCases(path string) ([]EvalCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cases []EvalCase
	if err := json.Unmarshal(data, &cases); err != nil {
		return nil, err
	}
	return cases, nil
}

// RunEval 逐个跑用例,返回 (通过数, 总数)。
func RunEval(ctx context.Context, llm ChatLLM, reg *Registry, cases []EvalCase, logger *Logger) (passed, total int) {
	for i, c := range cases {
		answer, err := runReAct(ctx, llm, reg, c.Question, "")
		ok := err == nil && strings.Contains(answer, c.MustContain)
		if ok {
			passed++
		}
		logger.Log("info", "eval_case", map[string]any{
			"index":        i,
			"question":     c.Question,
			"must_contain": c.MustContain,
			"passed":       ok,
			"answer":       truncateStr(answer, 100),
		})
	}
	return passed, len(cases)
}
