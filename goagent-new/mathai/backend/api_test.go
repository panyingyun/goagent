package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"mathai/pkg/llm"
)

func TestSessionChatAPI(t *testing.T) {
	t.Parallel()
	store, err := OpenStore(filepath.Join(t.TempDir(), "sessions.json"))
	if err != nil {
		t.Fatal(err)
	}
	client := &scriptedClient{responses: []llm.Response{
		{
			Message: llm.Message{
				Role: llm.RoleAssistant,
				ToolCalls: []llm.ToolCall{{
					ID: "call_1", Type: "function",
					Function: llm.FunctionCall{
						Name:      "apply_graph_patch",
						Arguments: `{"operation":"add_equation","equation":{"expression":"sin(x)"}}`,
					},
				}},
			},
		},
		{Message: llm.Message{Role: llm.RoleAssistant, Content: "已绘制正弦函数。"}},
	}}
	api := NewAPI(store, NewMathAgent(client), NewTraceStore(10), "", "", true)
	server := httptest.NewServer(api.Handler())
	defer server.Close()

	created := requestJSON[Session](t, http.MethodPost, server.URL+"/api/sessions", map[string]string{"title": "三角函数"}, http.StatusCreated)
	response := requestJSON[struct {
		Session Session `json:"session"`
		TraceID string  `json:"traceId"`
	}](t, http.MethodPost, server.URL+"/api/sessions/"+created.ID+"/chat", map[string]string{"message": "画 sin(x)"}, http.StatusOK)
	if len(response.Session.GraphState.Equations) != 1 || response.TraceID == "" {
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
