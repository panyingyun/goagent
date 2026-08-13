package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPClientComplete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer secret" {
			t.Fatalf("authorization = %q", got)
		}
		w.Header().Set("X-Request-Id", "req-1")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":2,"completion_tokens":1,"total_tokens":3}}`))
	}))
	defer srv.Close()
	c := NewHTTPClient("secret", srv.URL, "test-model")
	got, err := c.Complete(context.Background(), Request{Messages: []Message{{Role: RoleUser, Content: "hi"}}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Message.Content != "ok" || got.Usage.TotalTokens != 3 || got.RequestID != "req-1" {
		t.Fatalf("unexpected response: %+v", got)
	}
}

func TestHTTPClientPrefersDeepSeekTraceID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Request-Id", "generic-req")
		w.Header().Set("X-Ds-Trace-Id", "ds-trace-1")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer srv.Close()
	c := NewHTTPClient("secret", srv.URL, "test-model")
	got, err := c.Complete(context.Background(), Request{Messages: []Message{{Role: RoleUser, Content: "hi"}}})
	if err != nil {
		t.Fatal(err)
	}
	if got.RequestID != "ds-trace-1" {
		t.Fatalf("request id = %q, want DeepSeek trace id", got.RequestID)
	}
}

func TestHTTPClientSendsResponseFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ResponseFormat *ResponseFormat `json:"response_format"`
			Temperature    *float64        `json:"temperature"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.ResponseFormat == nil || body.ResponseFormat.Type != "json_object" {
			t.Fatalf("response_format = %+v", body.ResponseFormat)
		}
		if body.Temperature == nil || *body.Temperature != 0 {
			t.Fatalf("temperature = %v", body.Temperature)
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"{}"},"finish_reason":"stop"}],"usage":{}}`))
	}))
	defer srv.Close()
	c := NewHTTPClient("secret", srv.URL, "test-model")
	zero := 0.0
	_, err := c.Complete(context.Background(), Request{
		Messages:       []Message{{Role: RoleUser, Content: "hi"}},
		Temperature:    &zero,
		ResponseFormat: &ResponseFormat{Type: "json_object"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestHTTPClientRetryableError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "busy", http.StatusTooManyRequests)
	}))
	defer srv.Close()
	c := NewHTTPClient("secret", srv.URL, "test-model")
	_, err := c.Complete(context.Background(), Request{})
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) || !httpErr.Retryable() {
		t.Fatalf("expected retryable HTTPError, got %v", err)
	}
}

func TestHTTPClientPropagatesCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer srv.Close()
	c := NewHTTPClient("secret", srv.URL, "test-model")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := c.Complete(ctx, Request{})
	if err == nil {
		t.Fatal("expected cancellation error")
	}
}

