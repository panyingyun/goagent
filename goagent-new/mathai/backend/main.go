package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mathai/pkg/llm"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	})))

	evalOnly := flag.Bool("eval", false, "run testdata/eval-cases.json and exit")
	flag.Parse()

	client := llm.NewHTTPClientFromEnv()
	if *evalOnly {
		if err := runEvalCommand(client); err != nil {
			slog.Error("eval failed", "err", err)
			os.Exit(1)
		}
		return
	}

	store, err := OpenStore(envOr("MATHAI_DATA_FILE", "data/sessions.json"))
	if err != nil {
		slog.Error("open store failed", "err", err)
		os.Exit(1)
	}
	application := NewAPI(store, NewMathAgent(client), NewTraceStore(200), os.Getenv("MATHAI_API_TOKEN"), os.Getenv("MATHAI_ALLOW_ORIGIN"), client.Configured())
	server := &http.Server{
		Addr:              ":" + envOr("PORT", "8080"),
		Handler:           application.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      80 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		slog.Info("MathAI listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
	}
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
