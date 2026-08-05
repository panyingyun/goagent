package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testAPI() *API {
	return &API{Service: ChatFunc(func(context.Context, string) (string, error) { return "real service answer", nil }), Token: "secret", Ready: true, Timeout: time.Second}
}

func TestAPIAuthStrictJSONAndRealService(t *testing.T) {
	api := testAPI()
	request := httptest.NewRequest(http.MethodPost, "/v1/chat", strings.NewReader(`{"message":"hi"}`))
	request.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	api.Handler().ServeHTTP(w, request)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", w.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/v1/chat", strings.NewReader(`{"message":"hi","admin":true}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer secret")
	w = httptest.NewRecorder()
	api.Handler().ServeHTTP(w, request)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("unknown field status = %d", w.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/v1/chat", strings.NewReader(`{"message":"hi"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer secret")
	w = httptest.NewRecorder()
	api.Handler().ServeHTTP(w, request)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "real service answer") {
		t.Fatalf("response = %d %s", w.Code, w.Body.String())
	}
}

func TestReadyReturns503(t *testing.T) {
	api := testAPI()
	api.Ready = false
	w := httptest.NewRecorder()
	api.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", w.Code)
	}
}

