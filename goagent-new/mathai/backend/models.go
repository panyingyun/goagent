package main

import "time"

type Session struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Version    int64      `json:"version"`
	Messages   []Message  `json:"messages"`
	GraphState GraphState `json:"graphState"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type SessionSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Version   int64     `json:"version"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	TraceID   string    `json:"traceId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type GraphState struct {
	Equations []Equation     `json:"equations"`
	Viewport  Viewport       `json:"viewport"`
	Settings  GraphSettings  `json:"settings"`
	Analysis  *GraphAnalysis `json:"analysis,omitempty"`
}

type Equation struct {
	ID         string  `json:"id"`
	Expression string  `json:"expression"`
	Label      string  `json:"label"`
	Color      string  `json:"color"`
	Visible    bool    `json:"visible"`
	LineWidth  float64 `json:"lineWidth"`
}

type Viewport struct {
	XMin float64 `json:"xMin"`
	XMax float64 `json:"xMax"`
	YMin float64 `json:"yMin"`
	YMax float64 `json:"yMax"`
}

type GraphSettings struct {
	ShowGrid    bool `json:"showGrid"`
	ShowAxes    bool `json:"showAxes"`
	ShowLegend  bool `json:"showLegend"`
	SampleCount int  `json:"sampleCount"`
}

type GraphAnalysis struct {
	Summary    string   `json:"summary"`
	Highlights []string `json:"highlights"`
}

type GraphPatch struct {
	Operation  string          `json:"operation"`
	Equations  []EquationInput `json:"equations,omitempty"`
	Equation   *EquationInput  `json:"equation,omitempty"`
	EquationID string          `json:"equationId,omitempty"`
	Updates    *EquationUpdate `json:"updates,omitempty"`
	Viewport   *ViewportUpdate `json:"viewport,omitempty"`
	Settings   *SettingsUpdate `json:"settings,omitempty"`
	Analysis   *GraphAnalysis  `json:"analysis,omitempty"`
}

type EquationInput struct {
	Expression string   `json:"expression"`
	Label      string   `json:"label,omitempty"`
	Color      string   `json:"color,omitempty"`
	Visible    *bool    `json:"visible,omitempty"`
	LineWidth  *float64 `json:"lineWidth,omitempty"`
}

type EquationUpdate struct {
	Expression *string  `json:"expression,omitempty"`
	Label      *string  `json:"label,omitempty"`
	Color      *string  `json:"color,omitempty"`
	Visible    *bool    `json:"visible,omitempty"`
	LineWidth  *float64 `json:"lineWidth,omitempty"`
}

type ViewportUpdate struct {
	XMin *float64 `json:"xMin,omitempty"`
	XMax *float64 `json:"xMax,omitempty"`
	YMin *float64 `json:"yMin,omitempty"`
	YMax *float64 `json:"yMax,omitempty"`
}

type SettingsUpdate struct {
	ShowGrid    *bool `json:"showGrid,omitempty"`
	ShowAxes    *bool `json:"showAxes,omitempty"`
	ShowLegend  *bool `json:"showLegend,omitempty"`
	SampleCount *int  `json:"sampleCount,omitempty"`
}

type DollarPowerInput struct {
	Base     float64 `json:"base"`
	Exponent float64 `json:"exponent"`
}

type TraceRecord struct {
	ID        string    `json:"id"`
	SessionID string    `json:"sessionId"`
	CreatedAt time.Time `json:"createdAt"`
	Spans     any       `json:"spans"`
	Events    any       `json:"events"`
}
