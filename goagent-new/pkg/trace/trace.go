package trace

import (
	"context"
	"encoding/json"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"goagent-new/pkg/agent"
)

type Span struct {
	ID       string         `json:"id"`
	ParentID string         `json:"parent_id,omitempty"`
	Name     string         `json:"name"`
	Started  time.Time      `json:"started"`
	Ended    time.Time      `json:"ended"`
	Attrs    map[string]any `json:"attrs,omitempty"`
	Error    string         `json:"error,omitempty"`
}

type spanKey struct{}

type Recorder struct {
	mu      sync.Mutex
	nextID  atomic.Uint64
	spans   []*Span
	events  []agent.Event
}

func NewRecorder() *Recorder { return &Recorder{} }

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

func (r *Recorder) Record(_ context.Context, event agent.Event) {
	r.mu.Lock()
	r.events = append(r.events, event)
	r.mu.Unlock()
}

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

