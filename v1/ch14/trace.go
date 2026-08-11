package main

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

// Span 是一次调用的一段记录(类似 LangSmith 的 run / OpenTelemetry 的 span)。
type Span struct {
	ID       string
	ParentID string
	Name     string
	Start    time.Time
	End      time.Time
	Attrs    map[string]any
	Err      error
}

// Tracer 维护一个 span 树:Start 入栈,End 出栈,自动形成父子关系。
type Tracer struct {
	mu     sync.Mutex
	spans  []*Span
	active []*Span
}

func NewTracer() *Tracer {
	return &Tracer{}
}

func (t *Tracer) Start(name string) *Span {
	t.mu.Lock()
	defer t.mu.Unlock()
	parent := ""
	if len(t.active) > 0 {
		parent = t.active[len(t.active)-1].ID
	}
	s := &Span{
		ID:       newID(),
		ParentID: parent,
		Name:     name,
		Start:    time.Now(),
		Attrs:    map[string]any{},
	}
	t.spans = append(t.spans, s)
	t.active = append(t.active, s)
	return s
}

func (t *Tracer) End(s *Span) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s.End = time.Now()
	for i := len(t.active) - 1; i >= 0; i-- {
		if t.active[i] == s {
			t.active = append(t.active[:i], t.active[i+1:]...)
			break
		}
	}
}

// Export 把所有 span 输出成 JSON 数组(含父子关系、耗时、属性)。
func (t *Tracer) Export(w io.Writer) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]map[string]any, 0, len(t.spans))
	for _, s := range t.spans {
		e := map[string]any{
			"id":        s.ID,
			"parent_id": s.ParentID,
			"name":      s.Name,
			"duration":  s.End.Sub(s.Start).String(),
			"attrs":     s.Attrs,
		}
		if s.Err != nil {
			e["error"] = s.Err.Error()
		}
		out = append(out, e)
	}
	return json.NewEncoder(w).Encode(out)
}

// newID 用时间戳生成简易 ID(生产应使用 UUID)。
func newID() string {
	return time.Now().Format("150405.000000000")
}
