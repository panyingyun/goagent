package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"goagent-new/pkg/llm"
	"goagent-new/pkg/trace"
	"goagent-new/pkg/workspace"
)

const maxRequestBytes = 32 << 10

type api struct {
	app      *workspace.App
	token    string
	ready    bool
	limiter  *limiter
}

func main() {
	token := os.Getenv("AGENT_API_TOKEN")
	if token == "" {
		log.Fatal("AGENT_API_TOKEN must be configured")
	}
	apiKey, model := os.Getenv("LLM_API_KEY"), os.Getenv("LLM_MODEL")
	client := llm.NewHTTPClient(apiKey, envOr("LLM_BASE_URL", "https://api.deepseek.com"), model)
	root := envOr("WORKSPACE_DIR", filepath.FromSlash("testdata/workspace"))
	app, err := workspace.New(root, client, nil, trace.NewRecorder())
	if err != nil {
		log.Fatal(err)
	}
	h := &api{app: app, token: token, ready: apiKey != "" && model != "", limiter: newLimiter(30, time.Minute)}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, http.StatusOK, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /readyz", h.handleReady)
	mux.Handle("POST /v1/chat", h.authenticate(http.HandlerFunc(h.handleChat)))

	server := &http.Server{
		Addr:              ":" + envOr("PORT", "8080"),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      75 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		log.Printf("workspace server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server failed: %v", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func (a *api) handleReady(w http.ResponseWriter, _ *http.Request) {
	if !a.ready {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (a *api) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+a.token {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if !a.limiter.Allow(ip, time.Now()) {
			writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate limit exceeded"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *api) handleChat(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeJSON(w, http.StatusUnsupportedMediaType, map[string]string{"error": "Content-Type must be application/json"})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	var request struct {
		Message string `json:"message"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON request"})
		return
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request must contain one JSON object"})
		return
	}
	if strings.TrimSpace(request.Message) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message is required"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	answer, _, err := a.app.Chat(ctx, nil, request.Message)
	if err != nil {
		log.Printf("chat failed: %v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "agent request failed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"answer": answer})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

type limiter struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	clients map[string]bucket
	lastGC  time.Time
}

type bucket struct {
	started time.Time
	count   int
}

func newLimiter(max int, window time.Duration) *limiter {
	return &limiter{max: max, window: window, clients: make(map[string]bucket), lastGC: time.Now()}
}

func (l *limiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if now.Sub(l.lastGC) >= l.window {
		for client, b := range l.clients {
			if now.Sub(b.started) >= l.window {
				delete(l.clients, client)
			}
		}
		l.lastGC = now
	}
	b := l.clients[key]
	if b.started.IsZero() || now.Sub(b.started) >= l.window {
		b = bucket{started: now}
	}
	if b.count >= l.max {
		l.clients[key] = b
		return false
	}
	b.count++
	l.clients[key] = b
	return true
}
