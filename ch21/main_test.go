package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := newRateLimiter(2, time.Second)
	if !rl.Allow("ip1") || !rl.Allow("ip1") {
		t.Fatal("前两次应放行")
	}
	if rl.Allow("ip1") {
		t.Fatal("第三次应被拒绝")
	}
	if !rl.Allow("ip2") {
		t.Fatal("不同 IP 不应受连坐")
	}
}

func TestReadyHandler(t *testing.T) {
	// 未配置 Key -> not_ready
	rec := httptest.NewRecorder()
	handleReady("")(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	var out map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["status"] != "not_ready" {
		t.Fatalf("未配置 Key 应为 not_ready: %v", out)
	}

	// 已配置 Key -> ready
	rec = httptest.NewRecorder()
	handleReady("sk-test")(rec, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out["status"] != "ready" {
		t.Fatalf("已配置 Key 应为 ready: %v", out)
	}
}

func TestChatHandler(t *testing.T) {
	rl := newRateLimiter(5, time.Second)
	mux := http.NewServeMux()
	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if !rl.Allow(clientIP(r)) {
			http.Error(w, "限流", http.StatusTooManyRequests)
			return
		}
		writeJSON(w, map[string]any{"answer": "ok"})
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/chat", "application/json", strings.NewReader(`{"question":"hi"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("正常请求应 200,实际 %d", resp.StatusCode)
	}
}
