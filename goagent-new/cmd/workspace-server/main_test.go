package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goagent-new/pkg/llm"
	"goagent-new/pkg/workspace"
)

type directClient struct{}

func (directClient) Complete(context.Context, llm.Request) (llm.Response, error) {
	return llm.Response{Message: llm.Message{Role: llm.RoleAssistant, Content: "ok"}}, nil
}

func testAPI(t *testing.T) *api {
	t.Helper()
	app, err := workspace.New(t.TempDir(), directClient{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	return &api{app: app, token: "secret", ready: true, limiter: newLimiter(10, 1)}
}

func TestChatRequiresAuthenticationAndStrictJSON(t *testing.T) {
	a := testAPI(t)
	handler := a.authenticate(http.HandlerFunc(a.handleChat))
	unauthorized := httptest.NewRequest(http.MethodPost, "/v1/chat", strings.NewReader(`{"message":"hi"}`))
	unauthorized.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, unauthorized)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}

	bad := httptest.NewRequest(http.MethodPost, "/v1/chat", strings.NewReader(`{"message":"hi","extra":true}`))
	bad.Header.Set("Authorization", "Bearer secret")
	bad.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, bad)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestChatCallsAgent(t *testing.T) {
	a := testAPI(t)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat", strings.NewReader(`{"message":"hi"}`))
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	a.authenticate(http.HandlerFunc(a.handleChat)).ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"answer":"ok"`) {
		t.Fatalf("response = %d %s", w.Code, w.Body.String())
	}
}

func TestReadyUses503(t *testing.T) {
	a := testAPI(t)
	a.ready = false
	w := httptest.NewRecorder()
	a.handleReady(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", w.Code)
	}
}

