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
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxRequestBytes = 32 << 10

type API struct {
	store       *Store
	agent       ChatProcessor
	traces      *TraceStore
	token       string
	allowOrigin string
	ready       bool
	limiter     *rateLimiter
}

func NewAPI(store *Store, processor ChatProcessor, traces *TraceStore, token, allowOrigin string, ready bool) *API {
	return &API{store: store, agent: processor, traces: traces, token: token, allowOrigin: allowOrigin, ready: ready, limiter: newRateLimiter(60, time.Minute)}
}

func (a *API) Handler(staticDirectory string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", a.handleReady)
	mux.Handle("GET /api/sessions", a.protect(http.HandlerFunc(a.listSessions)))
	mux.Handle("POST /api/sessions", a.protect(http.HandlerFunc(a.createSession)))
	mux.Handle("GET /api/sessions/{id}", a.protect(http.HandlerFunc(a.getSession)))
	mux.Handle("DELETE /api/sessions/{id}", a.protect(http.HandlerFunc(a.deleteSession)))
	mux.Handle("POST /api/sessions/{id}/chat", a.protect(http.HandlerFunc(a.chat)))
	mux.Handle("GET /api/traces/{id}", a.protect(http.HandlerFunc(a.getTrace)))
	mux.Handle("/", spaHandler(staticDirectory))
	return a.cors(mux)
}

func (a *API) handleReady(w http.ResponseWriter, _ *http.Request) {
	if !a.ready {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready", "reason": "LLM_API_KEY and LLM_MODEL are required"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (a *API) listSessions(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"sessions": a.store.List()})
}

func (a *API) createSession(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Title string `json:"title"`
	}
	if err := readJSON(w, r, &request); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	session, err := a.store.Create(request.Title)
	if err != nil {
		log.Printf("create session: %v", err)
		writeAPIError(w, http.StatusInternalServerError, "could not create session")
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func (a *API) getSession(w http.ResponseWriter, r *http.Request) {
	session, err := a.store.Get(r.PathValue("id"))
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

func (a *API) deleteSession(w http.ResponseWriter, r *http.Request) {
	err := a.store.Delete(r.PathValue("id"))
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

func (a *API) chat(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Message string `json:"message"`
	}
	if err := readJSON(w, r, &request); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	request.Message = strings.TrimSpace(request.Message)
	if request.Message == "" || len(request.Message) > 4000 {
		writeAPIError(w, http.StatusBadRequest, "message must contain 1 to 4000 characters")
		return
	}
	session, err := a.store.Get(r.PathValue("id"))
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
	result, err := a.agent.Process(ctx, session, request.Message)
	if err != nil {
		log.Printf("MathAI chat failed session=%s: %v", session.ID, err)
		writeAPIError(w, http.StatusBadGateway, "MathAI request failed")
		return
	}
	now := time.Now().UTC()
	userMessage := Message{ID: newID("msg"), Role: "user", Content: request.Message, CreatedAt: now}
	assistantMessage := Message{ID: newID("msg"), Role: "assistant", Content: result.Answer, TraceID: result.TraceID, CreatedAt: time.Now().UTC()}
	updated, err := a.store.CommitChat(session.ID, session.Version, userMessage, assistantMessage, result.State)
	if errors.Is(err, ErrConflict) {
		writeAPIError(w, http.StatusConflict, "session changed during this request; retry with the latest state")
		return
	}
	if err != nil {
		log.Printf("commit chat: %v", err)
		writeAPIError(w, http.StatusInternalServerError, "could not save conversation")
		return
	}
	a.traces.Put(result.Trace)
	writeJSON(w, http.StatusOK, map[string]any{"session": updated, "answer": result.Answer, "traceId": result.TraceID})
}

func (a *API) getTrace(w http.ResponseWriter, r *http.Request) {
	record, ok := a.traces.Get(r.PathValue("id"))
	if !ok {
		writeAPIError(w, http.StatusNotFound, "trace not found or expired")
		return
	}
	writeJSON(w, http.StatusOK, record)
}

func (a *API) protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.token != "" && r.Header.Get("Authorization") != "Bearer "+a.token {
			writeAPIError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		key, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			key = r.RemoteAddr
		}
		if !a.limiter.Allow(key, time.Now()) {
			w.Header().Set("Retry-After", "60")
			writeAPIError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if a.allowOrigin != "" && origin == a.allowOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		}
		if r.Method == http.MethodOptions {
			if origin != "" && origin != a.allowOrigin {
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

func spaHandler(directory string) http.Handler {
	files := http.FileServer(http.Dir(directory))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		cleaned := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
		if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
			http.NotFound(w, r)
			return
		}
		if cleaned == "." {
			cleaned = "index.html"
		}
		candidate := filepath.Join(directory, cleaned)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			files.ServeHTTP(w, r)
			return
		}
		index := filepath.Join(directory, "index.html")
		if _, err := os.Stat(index); err != nil {
			writeAPIError(w, http.StatusNotFound, "frontend has not been built")
			return
		}
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, index)
	})
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
