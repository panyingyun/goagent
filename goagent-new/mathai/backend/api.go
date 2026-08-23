package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"log/slog"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const maxRequestBytes = 32 << 10

type API struct {
	store       *Store
	agent       *MathAgent
	traces      *TraceStore
	token       string
	allowOrigin string
	ready       bool
	limiter     *rateLimiter
}

// NewAPI 构造 HTTP API，绑定会话存储、聊天处理器与追踪存储。
func NewAPI(store *Store, agent *MathAgent, traces *TraceStore, token, allowOrigin string, ready bool) *API {
	return &API{store: store, agent: agent, traces: traces, token: token, allowOrigin: allowOrigin, ready: ready, limiter: newRateLimiter(60, time.Minute)}
}

// Handler 返回带 CORS 与鉴权中间件的路由。
func (api *API) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", api.handleReady)
	mux.Handle("GET /api/sessions", api.protect(http.HandlerFunc(api.listSessions)))
	mux.Handle("POST /api/sessions", api.protect(http.HandlerFunc(api.createSession)))
	mux.Handle("GET /api/sessions/{id}", api.protect(http.HandlerFunc(api.getSession)))
	mux.Handle("DELETE /api/sessions/{id}", api.protect(http.HandlerFunc(api.deleteSession)))
	mux.Handle("POST /api/sessions/{id}/chat", api.protect(http.HandlerFunc(api.chat)))
	mux.Handle("GET /api/traces/{id}", api.protect(http.HandlerFunc(api.getTrace)))
	return api.cors(mux)
}

func (api *API) handleReady(w http.ResponseWriter, _ *http.Request) {
	if !api.ready {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "reason": "DEEPSEEK_API_KEY is required"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (api *API) listSessions(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"sessions": api.store.List()})
}

func (api *API) createSession(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Title string `json:"title"`
	}
	if err := readJSON(w, r, &request); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	session, err := api.store.Create(request.Title)
	if err != nil {
		log.Printf("create session: %v", err)
		writeAPIError(w, http.StatusInternalServerError, "could not create session")
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (api *API) getSession(w http.ResponseWriter, r *http.Request) {
	session, err := api.store.Get(r.PathValue("id"))
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "could not read session")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (api *API) deleteSession(w http.ResponseWriter, r *http.Request) {
	err := api.store.Delete(r.PathValue("id"))
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "could not delete session")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) chat(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Message string `json:"message"`
	}
	if err := readJSON(w, r, &request); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	slog.Debug("chat request", "message", request.Message)

	request.Message = strings.TrimSpace(request.Message)
	if request.Message == "" || len(request.Message) > 4000 {
		writeAPIError(w, http.StatusBadRequest, "message must contain 1 to 4000 characters")
		return
	}
	slog.Debug("chat session", "id", r.PathValue("id"))
	session, err := api.store.Get(r.PathValue("id"))
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "could not read session")
		return
	}

	ctx, cancel := contextWithTimeout(r, 70*time.Second)
	defer cancel()
	result, err := api.agent.Process(ctx, session, request.Message)
	slog.Debug("chat result", "result", result)
	if err != nil {
		log.Printf("MathAI chat failed session=%s: %v", session.ID, err)
		writeAPIError(w, http.StatusBadGateway, "MathAI request failed")
		return
	}
	now := time.Now().UTC()
	userMessage := Message{ID: newID("msg"), Role: "user", Content: request.Message, CreatedAt: now}
	assistantMessage := Message{ID: newID("msg"), Role: "assistant", Content: result.Answer, TraceID: result.TraceID, CreatedAt: time.Now().UTC()}

	slog.Debug("chat messages", "user", request.Message, "assistant", result.Answer, "state", result.State)

	updated, err := api.store.CommitChat(session.ID, session.Version, userMessage, assistantMessage, result.State)
	slog.Debug("chat committed", "session", updated)
	if errors.Is(err, ErrConflict) {
		writeAPIError(w, http.StatusConflict, "session changed during this request; retry with the latest state")
		return
	}
	if err != nil {
		log.Printf("commit chat: %v", err)
		writeAPIError(w, http.StatusInternalServerError, "could not save conversation")
		return
	}
	api.traces.Put(result.Trace)
	writeJSON(w, http.StatusOK, map[string]any{"session": updated, "answer": result.Answer, "traceId": result.TraceID})
}

func (api *API) getTrace(w http.ResponseWriter, r *http.Request) {
	record, ok := api.traces.Get(r.PathValue("id"))
	if !ok {
		writeAPIError(w, http.StatusNotFound, "trace not found or expired")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (api *API) protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if api.token != "" && r.Header.Get("Authorization") != "Bearer "+api.token {
			writeAPIError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		key, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			key = r.RemoteAddr
		}
		if !api.limiter.Allow(key, time.Now()) {
			w.Header().Set("Retry-After", "60")
			writeAPIError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (api *API) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if api.allowOrigin != "" && origin == api.allowOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			if origin != "" && origin != api.allowOrigin {
				writeAPIError(w, http.StatusForbidden, "origin is not allowed")
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func readJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return errors.New("Content-Type must be application/json")
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return errors.New("invalid JSON request")
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("request must contain one JSON object")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeAPIError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

type rateLimiter struct {
	mu      sync.Mutex
	maximum int
	window  time.Duration
	clients map[string]rateBucket
}

type rateBucket struct {
	started time.Time
	count   int
}

func newRateLimiter(maximum int, window time.Duration) *rateLimiter {
	return &rateLimiter{maximum: maximum, window: window, clients: make(map[string]rateBucket)}
}

func (l *rateLimiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	bucket := l.clients[key]
	if bucket.started.IsZero() || now.Sub(bucket.started) >= l.window {
		bucket = rateBucket{started: now}
	}
	if bucket.count >= l.maximum {
		l.clients[key] = bucket
		return false
	}
	bucket.count++
	l.clients[key] = bucket
	return true
}

func contextWithTimeout(r *http.Request, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), timeout)
}
