package main

import (
	_ "embed"
	"encoding/json"
	"strings"
)

// systemPrompt 来自设计稿 §11.1,以独立文件管理(呼应 ch05 的 SKILL 思想:
// 提示词是资产,不是代码)。go:embed 让它随二进制打包,运行时不依赖磁盘路径。
//
//go:embed prompt.txt
var systemPrompt string

// buildUserPrompt 构造 user 消息(设计稿 §11.2)。
func buildUserPrompt(userMessage string, gs *GraphState, recent []llmMessage) string {
	p := map[string]any{
		"userMessage": userMessage,
		"currentGraphState": map[string]any{
			"equations": gs.Equations,
			"viewport":  gs.Viewport,
		},
		"recentMessages": recent,
	}
	b, _ := json.Marshal(p)
	return string(b)
}

// ParseStructuredResult 解析 + 修复 + 校验 DeepSeek 返回(设计稿 §13)。
func ParseStructuredResult(raw string) (*LLMStructuredResult, error) {
	clean := stripCodeFence(raw)
	var res LLMStructuredResult
	if err := json.Unmarshal([]byte(clean), &res); err != nil {
		if obj := extractJSONObject(clean); obj != "" {
			err = json.Unmarshal([]byte(obj), &res)
		}
		if err != nil {
			return nil, err
		}
	}
	// 校验 intent
	if !validIntents[res.Intent] {
		res.Intent = "unknown"
		res.Error = "非法 intent"
	}
	// 方程默认值补齐
	for i := range res.Equations {
		e := &res.Equations[i]
		if e.Type == "" {
			e.Type = "function"
		}
		if e.NormalizedExpression == "" && e.Expression != "" {
			e.NormalizedExpression = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(e.Expression), "y ="))
		}
		if e.Expression == "" && e.NormalizedExpression != "" {
			e.Expression = "y = " + e.NormalizedExpression
		}
		if e.Label == "" {
			e.Label = "y = " + e.NormalizedExpression
		}
		if e.Color == "" {
			e.Color = palette[i%len(palette)]
		}
		if e.LineWidth == 0 {
			e.LineWidth = 2
		}
		if e.Expression != "" {
			e.Visible = true // 新方程默认可见(visible 缺失时 bool 零值无法区分)
		}
	}
	return &res, nil
}

func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if i := strings.Index(s[3:], "\n"); i >= 0 {
			s = s[3+i+1:]
		} else {
			s = s[3:]
		}
	}
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
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
