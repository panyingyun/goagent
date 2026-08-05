package main

import (
	"context"
	"testing"

	"goagent-new/pkg/trace"
)

func TestRecordDemoCreatesParentedSpansAndUsageEvents(t *testing.T) {
	recorder := trace.NewRecorder()
	recordDemo(context.Background(), recorder)
	spans, events := recorder.Snapshot()
	if len(spans) != 4 || len(events) != 2 {
		t.Fatalf("spans=%+v events=%+v", spans, events)
	}
	if spans[1].ParentID != spans[0].ID || events[0].InputTokens != 20 || spans[3].Error == "" {
		t.Fatalf("trace lost important data: spans=%+v events=%+v", spans, events)
	}
}

