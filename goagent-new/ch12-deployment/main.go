package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"goagent-new/pkg/llm"
	"goagent-new/pkg/workspace"
)

func main() {
	token := os.Getenv("AGENT_API_TOKEN")
	if token == "" {
		log.Fatal("AGENT_API_TOKEN must be configured")
	}
	apiKey, model := os.Getenv("LLM_API_KEY"), os.Getenv("LLM_MODEL")
	client := llm.NewHTTPClient(apiKey, envOr("LLM_BASE_URL", "https://api.deepseek.com"), model)
	app, err := workspace.New(envOr("WORKSPACE_DIR", filepath.FromSlash("testdata/workspace")), client, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	api := &API{
		Service: ChatFunc(func(ctx context.Context, message string) (string, error) {
			answer, _, err := app.Chat(ctx, nil, message)
			return answer, err
		}),
		Token: token, Ready: apiKey != "" && model != "", Timeout: 60 * time.Second,
	}
	server := &http.Server{
		Addr: ":" + envOr("PORT", "8080"), Handler: api.Handler(),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 75 * time.Second, IdleTimeout: 60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		log.Printf("deployment lesson listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server failed: %v", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

