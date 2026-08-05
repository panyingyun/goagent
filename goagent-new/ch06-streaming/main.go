package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goagent-new/pkg/stream"
)

func newStreamHandler(delay time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		text := r.URL.Query().Get("q")
		if text == "" {
			http.Error(w, "q is required", http.StatusBadRequest)
			return
		}
		events := make(chan stream.Event)
		go produce(r.Context(), text, delay, events)
		_ = stream.Relay(r.Context(), w, events)
	})
}

func produce(ctx context.Context, text string, delay time.Duration, events chan<- stream.Event) {
	defer close(events)
	for _, char := range []rune(text) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		select {
		case events <- stream.Event{Type: "delta", Data: string(char)}:
		case <-ctx.Done():
			return
		}
	}
	select {
	case events <- stream.Event{Type: "done"}:
	case <-ctx.Done():
	}
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("GET /stream", newStreamHandler(80*time.Millisecond))
	server := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		log.Println("open http://localhost:8080/stream?q=hello")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server failed: %v", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

