package trace

import (
	"context"
	"testing"
)

func TestParentPropagationIsContextScoped(t *testing.T) {
	r := NewRecorder()
	ctx, endRoot := r.Start(context.Background(), "root", nil)
	_, endChild := r.Start(ctx, "child", nil)
	endChild(nil)
	endRoot(nil)
	spans, _ := r.Snapshot()
	if len(spans) != 2 || spans[1].ParentID != spans[0].ID {
		t.Fatalf("unexpected spans: %+v", spans)
	}
	_, endOther := r.Start(context.Background(), "other", nil)
	endOther(nil)
	spans, _ = r.Snapshot()
	if spans[2].ParentID != "" {
		t.Fatal("unrelated request inherited a parent span")
	}
}

