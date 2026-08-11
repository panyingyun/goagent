package main

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestSpanTree(t *testing.T) {
	tr := NewTracer()
	root := tr.Start("root")
	child := tr.Start("child")
	tracerEnd(tr, child)
	tracerEnd(tr, root)

	var buf bytes.Buffer
	if err := tr.Export(&buf); err != nil {
		t.Fatal(err)
	}
	var spans []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &spans); err != nil {
		t.Fatal(err)
	}
	if len(spans) != 2 {
		t.Fatalf("应有 2 个 span,实际 %d", len(spans))
	}
	childSpan := spans[1]
	if childSpan["parent_id"] != spans[0]["id"] {
		t.Fatalf("子 span 的 parent 应为 root: %v", childSpan)
	}
}

func tracerEnd(tr *Tracer, s *Span) {
	time.Sleep(time.Millisecond)
	tr.End(s)
}

func TestMetrics(t *testing.T) {
	m := NewMetrics()
	m.Inc("a")
	m.Inc("a")
	m.Observe("latency", 100)
	m.Observe("latency", 300)
	s := m.Snapshot()
	if s["counter_a"] != int64(2) {
		t.Fatalf("计数错误: %v", s)
	}
	if s["hist_latency_avg"] != 200.0 {
		t.Fatalf("平均耗时错误: %v", s)
	}
}
