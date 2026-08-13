package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

type fakeProcessor struct{}

func (fakeProcessor) Process(_ context.Context, session Session, _ string) (ChatResult, error) {
	state, err := ApplyGraphPatch(session.GraphState, GraphPatch{Operation: "add_equation", Equation: &EquationInput{Expression: "sin(x)"}})
	if err != nil {
		return ChatResult{}, err
	}
	traceID := "trace_test"
	return ChatResult{Answer: "已绘制正弦函数。", State: state, TraceID: traceID, Trace: TraceRecord{ID: traceID, SessionID: session.ID, CreatedAt: time.Now().UTC(), Spans: []any{}, Events: []any{}}}, nil
}

func TestSessionChatAPI(t *testing.T) {
	t.Parallel()
	store, err := OpenStore(filepath.Join(t.TempDir(), "sessions.json"))
	if err != nil {
		t.Fatal(err)
	}
	api := NewAPI(store, fakeProcessor{}, NewTraceStore(10), "", "", true)
	server := httptest.NewServer(api.Handler())
	defer server.Close()

	created := requestJSON[Session](t, http.MethodPost, server.URL+"/api/sessions", map[string]string{"title": "三角函数"}, http.StatusCreated)
	response := requestJSON[struct {
		Session Session `json:"session"`
		TraceID string  `json:"traceId"`
	}](t, http.MethodPost, server.URL+"/api/sessions/"+created.ID+"/chat", map[string]string{"message": "画 sin(x)"}, http.StatusOK)
	if len(response.Session.GraphState.Equations) != 1 || response.TraceID != "trace_test" {
		t.Fatalf("unexpected response: %+v", response)
	}
	trace := requestJSON[TraceRecord](t, http.MethodGet, server.URL+"/api/traces/"+response.TraceID, nil, http.StatusOK)
	if trace.SessionID != created.ID {
		t.Fatalf("unexpected trace: %+v", trace)
	}
}

func requestJSON[T any](t *testing.T, method, url string, body any, wantStatus int) T {
	t.Helper()
	var payload []byte
	if body != nil {
		payload, _ = json.Marshal(body)
	}
	request, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != wantStatus {
		t.Fatalf("status = %d, want %d", response.StatusCode, wantStatus)
	}
	var decoded T
	if response.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
			t.Fatal(err)
		}
	}
	return decoded
}
