package main

import (
	"encoding/json"
	"time"
)

// Run 对应 LangSmith 的 run:一次带父子关系的执行记录。
type Run struct {
	Name     string         `json:"name"`
	Attrs    map[string]any `json:"attrs"`
	Error    string         `json:"error,omitempty"`
	Start    time.Time      `json:"start"`
	End      time.Time      `json:"end"`
	Children []*Run         `json:"children,omitempty"`
	parent   *Run
}

// Tracer 维护 run 树:StartRun 入栈,EndRun 出栈。
type Tracer struct {
	root  *Run
	stack []*Run
}

func NewTracer() *Tracer {
	return &Tracer{}
}

func (t *Tracer) StartRun(name string) *Run {
	r := &Run{Name: name, Start: time.Now(), Attrs: map[string]any{}}
	if len(t.stack) > 0 {
		parent := t.stack[len(t.stack)-1]
		r.parent = parent
		parent.Children = append(parent.Children, r)
	} else {
		t.root = r
	}
	t.stack = append(t.stack, r)
	return r
}

func (t *Tracer) EndRun(r *Run) {
	r.End = time.Now()
	for i := len(t.stack) - 1; i >= 0; i-- {
		if t.stack[i] == r {
			t.stack = append(t.stack[:i], t.stack[i+1:]...)
			break
		}
	}
}

func (t *Tracer) FailRun(r *Run, err error) {
	r.Error = err.Error()
	t.EndRun(r)
}

// Export 输出嵌套的 run 树 JSON(和 LangSmith 的层级视图同构)。
func (t *Tracer) Export() map[string]any {
	if t.root == nil {
		return map[string]any{}
	}
	var toMap func(r *Run) map[string]any
	toMap = func(r *Run) map[string]any {
		m := map[string]any{
			"name":   r.Name,
			"attrs":  r.Attrs,
			"start":  r.Start.Format(time.RFC3339),
			"end":    r.End.Format(time.RFC3339),
			"tokens": r.Attrs["tokens"],
		}
		if r.Error != "" {
			m["error"] = r.Error
		}
		if len(r.Children) > 0 {
			kids := make([]map[string]any, 0, len(r.Children))
			for _, c := range r.Children {
				kids = append(kids, toMap(c))
			}
			m["children"] = kids
		}
		return m
	}
	return toMap(t.root)
}

// ExportJSON 输出紧凑 JSON 字符串。
func (t *Tracer) ExportJSON() string {
	b, _ := json.MarshalIndent(t.Export(), "", "  ")
	return string(b)
}
