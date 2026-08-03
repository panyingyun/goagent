package main

import (
	"encoding/json"
	"fmt"
)

// 数据模型与设计稿 §6 一一对应。

type Session struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Messages   []Message  `json:"messages"`
	GraphState GraphState `json:"graphState"`
	IsFavorite bool       `json:"isFavorite"`
	CreatedAt  string     `json:"createdAt"`
	UpdatedAt  string     `json:"updatedAt"`
}

type Message struct {
	ID               string               `json:"id"`
	Role             string               `json:"role"`
	Content          string               `json:"content"`
	StructuredResult *LLMStructuredResult `json:"structuredResult,omitempty"`
	CreatedAt        string               `json:"createdAt"`
	Status           string               `json:"status"`
}

type GraphState struct {
	Equations []EquationItem `json:"equations"`
	Viewport  Viewport       `json:"viewport"`
	Settings  GraphSettings  `json:"settings"`
	Analysis  *GraphAnalysis `json:"analysis,omitempty"`
}

type EquationItem struct {
	ID                   FlexID  `json:"id"`
	Type                 string  `json:"type"`
	Expression           string  `json:"expression"`
	NormalizedExpression string  `json:"normalizedExpression"`
	Label                string  `json:"label"`
	Color                string  `json:"color"`
	Visible              bool    `json:"visible"`
	LineWidth            float64 `json:"lineWidth"`
}

type Viewport struct {
	XMin float64 `json:"xMin"`
	XMax float64 `json:"xMax"`
	YMin float64 `json:"yMin"`
	YMax float64 `json:"yMax"`
}

// ViewportUpdate 用指针字段表示"只更新用户提到的部分"。
type ViewportUpdate struct {
	XMin *float64 `json:"xMin"`
	XMax *float64 `json:"xMax"`
	YMin *float64 `json:"yMin"`
	YMax *float64 `json:"yMax"`
}

type GraphSettings struct {
	ShowGrid    bool `json:"showGrid"`
	ShowAxis    bool `json:"showAxis"`
	ShowLegend  bool `json:"showLegend"`
	SampleCount int  `json:"sampleCount"`
}

// EquationUpdate 是 update_equation 意图的部分更新(指针区分"没给"和"给 0")。
type EquationUpdate struct {
	Color                *string  `json:"color"`
	LineWidth            *float64 `json:"lineWidth"`
	Visible              *bool    `json:"visible"`
	Label                *string  `json:"label"`
	Expression           *string  `json:"expression"`
	NormalizedExpression *string  `json:"normalizedExpression"`
}

type KeyPoint struct {
	Label string  `json:"label"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

type GraphAnalysis struct {
	// 字段故意放宽为 any:模型经常输出格式不一的字段
	// (例如 zeros 给字符串 "x = kπ"),不能因此解析失败。
	FunctionType string `json:"functionType"`
	KeyPoints    any    `json:"keyPoints"`
	Monotonicity any    `json:"monotonicity"`
	Zeros        any    `json:"zeros"`
	Symmetry     string `json:"symmetry"`
	Asymptotes   any    `json:"asymptotes"`
	Description  string `json:"description"`
}

// LLMStructuredResult 是 DeepSeek 必须返回的结构化 JSON。
type LLMStructuredResult struct {
	Intent           string          `json:"intent"`
	Equations        []EquationItem  `json:"equations"`
	Viewport         *ViewportUpdate `json:"viewport"`
	TargetEquationID FlexID          `json:"targetEquationId"`
	Updates          *EquationUpdate `json:"updates"`
	Explanation      string          `json:"explanation"`
	Analysis         *GraphAnalysis  `json:"analysis"`
	Error            string          `json:"error"`
}

// FlexID 兼容模型把 id 输出成数字的情况(例如 "id": 1)。
// JSON 里始终以字符串形式输出。
type FlexID string

func (f *FlexID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*f = FlexID(s)
		return nil
	}
	var n float64
	if err := json.Unmarshal(b, &n); err == nil {
		*f = FlexID(fmt.Sprintf("%.0f", n))
		return nil
	}
	return fmt.Errorf("id 必须是字符串或数字: %s", b)
}

func defaultGraphState() GraphState {
	return GraphState{
		Equations: []EquationItem{},
		Viewport:  Viewport{XMin: -10, XMax: 10, YMin: -10, YMax: 10},
		Settings:  GraphSettings{ShowGrid: true, ShowAxis: true, ShowLegend: true, SampleCount: 1000},
	}
}

var validIntents = map[string]bool{
	"plot": true, "add_equation": true, "update_equation": true,
	"remove_equation": true, "update_viewport": true, "analyze": true,
	"explain": true, "unknown": true,
}

var palette = []string{"#2563eb", "#ef4444", "#22c55e", "#a855f7", "#f97316"}
