package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestDecodeJSONStrict(t *testing.T) {
	var request struct {
		Message string `json:"message"`
	}
	if err := DecodeJSONStrict(strings.NewReader(`{"message":"hello"}`), 64, &request); err != nil {
		t.Fatal(err)
	}
	if request.Message != "hello" {
		t.Fatalf("message = %q", request.Message)
	}
	if err := DecodeJSONStrict(strings.NewReader(`{"message":"hello","admin":true}`), 64, &request); err == nil {
		t.Fatal("unknown field should be rejected")
	}
	if err := DecodeJSONStrict(strings.NewReader(`{"message":"hello"} {}`), 64, &request); err == nil {
		t.Fatal("multiple JSON values should be rejected")
	}
	if err := DecodeJSONStrict(strings.NewReader(`{"message":"this body is too large"}`), 10, &request); err == nil {
		t.Fatal("oversized body should be rejected")
	}
}

func TestFetchTextUsesContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte("late response"))
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := FetchText(ctx, server.Client(), server.URL, 1024); err == nil {
		t.Fatal("expected timeout or cancellation")
	}
}

func TestFetchTextLimitsResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 100)))
	}))
	defer server.Close()
	if _, err := FetchText(context.Background(), server.Client(), server.URL, 10); err == nil {
		t.Fatal("oversized response should be rejected")
	}
}

func TestRunWorkersIsBoundedAndOrdered(t *testing.T) {
	var active atomic.Int32
	var maximum atomic.Int32
	results, err := RunWorkers(context.Background(), 2, []int{5, 4, 3, 2, 1}, func(ctx context.Context, value int) (int, error) {
		current := active.Add(1)
		defer active.Add(-1)
		for {
			old := maximum.Load()
			if current <= old || maximum.CompareAndSwap(old, current) {
				break
			}
		}
		select {
		case <-time.After(time.Duration(value) * time.Millisecond):
			return value * 10, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if maximum.Load() > 2 {
		t.Fatalf("maximum concurrency = %d", maximum.Load())
	}
	want := []int{50, 40, 30, 20, 10}
	for i := range want {
		if results[i] != want[i] {
			t.Fatalf("results = %v", results)
		}
	}
}

