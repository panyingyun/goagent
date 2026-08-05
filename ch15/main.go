// 课时15:部署与运维
//
// 把 Agent 变成可交付的服务:
// - /health 存活检查, /ready 就绪检查(依赖 API Key 是否配置)
// - /chat 带限流(每 IP 每秒最多 N 次)
// - 优雅关闭(收到退出信号后停止接收新请求)
// - 密钥只从环境变量读取,不进代码
//
// 运行(已配置 DEEPSEEK_API_KEY):
//
//	go run .
//
// 配合 Dockerfile: docker build -t goagent . && docker run -p 8080:8080 -e DEEPSEEK_API_KEY=... goagent
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	port := envOr("PORT", "8080")
	rl := newRateLimiter(5, time.Second) // 每 IP 每秒 5 次

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"status": "ok"})
	})
	mux.HandleFunc("/ready", handleReady(apiKey))
	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !rl.Allow(ip) {
			http.Error(w, "请求过于频繁,请稍后再试", http.StatusTooManyRequests)
			return
		}
		if apiKey == "" {
			http.Error(w, "服务未就绪:缺少 DEEPSEEK_API_KEY", http.StatusServiceUnavailable)
			return
		}
		var req struct {
			Question string `json:"question"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Question == "" {
			http.Error(w, "question 必填", http.StatusBadRequest)
			return
		}
		// 演示:真实项目这里调用 Agent
		writeJSON(w, map[string]any{"answer": "已收到: " + req.Question, "model_hint": "deepseek-v4-flash"})
	})

	srv := &http.Server{Addr: ":" + port, Handler: mux}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Println("服务启动:", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintln(os.Stderr, "服务异常退出:", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	fmt.Println("收到退出信号,正在优雅关闭...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Fprintln(os.Stderr, "关闭超时:", err)
	}
	fmt.Println("已关闭")
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// handleReady 就绪检查:依赖配置齐全才算 ready。
func handleReady(apiKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if apiKey == "" {
			writeJSON(w, map[string]any{"status": "not_ready", "reason": "DEEPSEEK_API_KEY 未配置"})
			return
		}
		writeJSON(w, map[string]any{"status": "ready"})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// rateLimiter 按 IP 分桶:每个 IP 在窗口内最多 max 次。
type rateLimiter struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	clients map[string]*clientBucket
}

type clientBucket struct {
	started time.Time
	count   int
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	return &rateLimiter{max: max, window: window, clients: map[string]*clientBucket{}}
}

func (r *rateLimiter) Allow(ip string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.clients[ip]
	if !ok {
		b = &clientBucket{started: time.Now()}
		r.clients[ip] = b
	}
	if time.Since(b.started) > r.window {
		b.started = time.Now()
		b.count = 0
	}
	if b.count >= r.max {
		return false
	}
	b.count++
	return true
}
