package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Judge 用另一个 LLM 给回答打分(LLM-as-Judge),解决"主观题无法断言"的问题。
type Judge struct {
	LLM      ChatLLM
	Criteria []string
}

func (j *Judge) Score(ctx context.Context, question, answer string) (float64, string, error) {
	prompt := "你是评分员。按以下标准给回答打分(0~10 分),只输出 JSON:" +
		`{"score": 0~10, "reason": "一句话理由"}` + "\n\n评分标准:\n" +
		strings.Join(j.Criteria, "\n") + "\n\n问题: " + question + "\n回答: " + answer
	reply, err := j.LLM.Chat(ctx, []Message{{Role: "user", Content: prompt}})
	if err != nil {
		return 0, "", err
	}
	var out struct {
		Score  float64 `json:"score"`
		Reason string  `json:"reason"`
	}
	jsonPart := extractJSONObject(reply.Content)
	if jsonPart == "" {
		return 0, "", fmt.Errorf("评分回复里没有 JSON: %q", reply.Content)
	}
	if err := json.Unmarshal([]byte(jsonPart), &out); err != nil {
		return 0, "", fmt.Errorf("解析评分失败: %w", err)
	}
	if out.Score < 0 {
		out.Score = 0
	}
	if out.Score > 10 {
		out.Score = 10
	}
	return out.Score, out.Reason, nil
}

func extractJSONObject(s string) string {
	start := strings.IndexByte(s, '{')
	if start < 0 {
		return ""
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}
