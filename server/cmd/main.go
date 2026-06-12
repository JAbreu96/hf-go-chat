// Command hf-go-chat starts the chat API server.
// Run with: go run ./cmd/main.go
// https://pkg.go.dev/net/http
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joelchrist/hf-go-chat/internal/config"
	"github.com/joelchrist/hf-go-chat/internal/handler"
	"github.com/joelchrist/hf-go-chat/internal/hf"
	"github.com/joelchrist/hf-go-chat/internal/middleware"
	"github.com/joelchrist/hf-go-chat/internal/rag"
	"github.com/joelchrist/hf-go-chat/internal/tools"
)

func main() {
	// slog is Go's structured logging package (added in Go 1.21).
	// Structured = key-value pairs alongside the message, easy to parse in log aggregators.
	// https://pkg.go.dev/log/slog
	slog.Info("starting hf-go-chat server")

	// Load returns (Config, error) — two return values.
	// := declares both cfg and err as new variables in this function scope.
	// https://go.dev/ref/spec#Short_variable_declarations
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "err", err)
		os.Exit(1)
	}

	// Load the SQuAD dataset from the HF Datasets Server into memory.
	// context.Background() is the root context — use it when there is no parent context.
	// https://pkg.go.dev/context#Background
	slog.Info("loading dataset", "name", cfg.DatasetName, "limit", cfg.DatasetLimit)
	rows, err := rag.Load(context.Background(), cfg.DatasetName, cfg.DatasetCfg, cfg.DatasetLimit)
	if err != nil {
		slog.Error("dataset load failed", "err", err)
		os.Exit(1)
	}
	slog.Info("dataset loaded", "rows", len(rows))

	// Wire up dependencies.
	hfClient := hf.NewClient(cfg.HFToken, cfg.Model)
	executor := tools.NewExecutor(cfg.BraveKey)
	chatHandler := handler.NewChat(hfClient, executor, rows)

	// http.NewServeMux creates a fresh request multiplexer (router).
	// The pattern "POST /api/chat" is a Go 1.22 enhanced pattern that matches
	// only POST requests to that path. https://pkg.go.dev/net/http#ServeMux
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)
	mux.Handle("POST /api/chat", chatHandler)

	// Wrap the entire mux with CORS middleware.
	// middleware.CORS(mux) returns a new http.Handler that sets CORS headers
	// before delegating to mux — the classic middleware chain pattern.
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.CORS(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second, // long for streaming responses
		IdleTimeout:  60 * time.Second,
	}

	// Goroutine — go func() launches a new goroutine: a lightweight thread managed
	// by the Go runtime (not the OS). The IIFE (immediately-invoked function expression)
	// pattern runs an anonymous function in the new goroutine.
	// https://go.dev/tour/concurrency/1
	go func() {
		slog.Info("listening", "addr", srv.Addr)
		// ListenAndServe blocks until the server stops.
		// http.ErrServerClosed is returned when Shutdown() is called — that's not an error.
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown: block the main goroutine until a termination signal arrives.
	//
	// make(chan os.Signal, 1) creates a buffered channel with capacity 1.
	// Buffered means signal.Notify can send without a receiver ready — capacity 1
	// ensures no signal is dropped if we're slow to receive.
	// https://go.dev/tour/concurrency/3
	quit := make(chan os.Signal, 1)
	// signal.Notify routes SIGINT (Ctrl-C) and SIGTERM (kill / Docker stop) to quit.
	// https://pkg.go.dev/os/signal#Notify
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// <-quit blocks the main goroutine until a value is received on the channel.
	// When a signal arrives, execution continues to shutdown.
	// https://go.dev/tour/concurrency/2
	<-quit
	slog.Info("shutting down")

	// Give in-flight requests up to 10 seconds to finish before hard-closing.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel() ensures the context's resources are freed even if Shutdown returns early.
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
}
