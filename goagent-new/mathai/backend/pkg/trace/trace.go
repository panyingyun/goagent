package trace

import (
	"context"
	"encoding/json"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"mathai/pkg/agent"
)

// Span 是一次命名操作的时间区间记录(可嵌套 parent)。
type Span struct {
	ID       string         `json:"id"`
	ParentID string         `json:"parent_id,omitempty"`
	Name     string         `json:"name"`
	Started  time.Time      `json:"started"`
	Ended    time.Time      `json:"ended"`
	Attrs    map[string]any `json:"attrs,omitempty"`
	Error    string         `json:"error,omitempty"`
}

// spanKey 用于把当前 Span 挂到 context 上。
type spanKey struct{}

// Recorder 收集 Span 与 agent.Event,支持快照与 JSON 导出。
type Recorder struct {
	mu     sync.Mutex
	nextID atomic.Uint64
	spans  []*Span
	events []agent.Event
}

// NewRecorder 创建空的追踪记录器。
func NewRecorder() *Recorder { return &Recorder{} }

// Start 开始一个命名 Span，返回带父 Span 的 context 以及结束回调。
func (r *Recorder) Start(ctx context.Context, name string, attrs map[string]any) (context.Context, func(error)) {
	span := &Span{ID: strconv.FormatUint(r.nextID.Add(1), 10), Name: name, Started: time.Now().UTC(), Attrs: cloneMap(attrs)}
	if parent, ok := ctx.Value(spanKey{}).(*Span); ok {
		span.ParentID = parent.ID
	}
	r.mu.Lock()
	r.spans = append(r.spans, span)
	r.mu.Unlock()
	child := context.WithValue(ctx, spanKey{}, span)
	var once sync.Once
	return child, func(err error) {
		once.Do(func() {
			r.mu.Lock()
			span.Ended = time.Now().UTC()
			if err != nil {
				span.Error = err.Error()
			}
			r.mu.Unlock()
		})
	}
}

// Record 实现 agent.Observer，追加一条运行时事件。
func (r *Recorder) Record(_ context.Context, event agent.Event) {
	r.mu.Lock()
	r.events = append(r.events, event)
	r.mu.Unlock()
}

// Snapshot 返回当前 Span 与事件的副本。
func (r *Recorder) Snapshot() ([]Span, []agent.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	spans := make([]Span, len(r.spans))
	for i, span := range r.spans {
		spans[i] = *span
		spans[i].Attrs = cloneMap(span.Attrs)
	}
	events := append([]agent.Event(nil), r.events...)
	return spans, events
}

// Export 将快照编码为 JSON 写入 w。
func (r *Recorder) Export(w io.Writer) error {
	spans, events := r.Snapshot()
	return json.NewEncoder(w).Encode(map[string]any{"spans": spans, "events": events})
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

