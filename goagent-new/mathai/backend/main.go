package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"goagent-new/pkg/llm"
)

func main() {
	apiKey, model := os.Getenv("LLM_API_KEY"), os.Getenv("LLM_MODEL")
	store, err := OpenStore(envOr("MATHAI_DATA_FILE", "data/sessions.json"))
	if err != nil {
		log.Fatal(err)
	}
	client := llm.NewHTTPClient(apiKey, envOr("LLM_BASE_URL", "https://api.deepseek.com"), model)
	application := NewAPI(store, NewMathAgent(client), NewTraceStore(200), os.Getenv("MATHAI_API_TOKEN"), os.Getenv("MATHAI_ALLOW_ORIGIN"), apiKey != "" && model != "")
	server := &http.Server{
		Addr: ":" + envOr("PORT", "8080"), Handler: application.Handler(envOr("MATHAI_STATIC_DIR", "frontend/dist")),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 80 * time.Second,
		IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		log.Printf("MathAI listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server failed: %v", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
