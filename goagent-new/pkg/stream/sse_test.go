package stream

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRelayWritesAllFrames(t *testing.T) {
	events := make(chan Event, 2)
	events <- Event{Type: "delta", Data: "你\n好"}
	events <- Event{Type: "done"}
	close(events)
	w := httptest.NewRecorder()
	if err := Relay(context.Background(), w, events); err != nil {
		t.Fatal(err)
	}
	body := w.Body.String()
	if !strings.Contains(body, "event: delta") || !strings.Contains(body, `你\n好`) || !strings.Contains(body, "event: done") {
		t.Fatalf("unexpected SSE body: %q", body)
	}
}

