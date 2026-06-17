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

// Packages available for your implementation — blank-identifier pins keep unused imports legal.
var (
	_ = context.Background  // root context with no parent — https://pkg.go.dev/context#Background
	_ = context.WithTimeout // context with a deadline — https://pkg.go.dev/context#WithTimeout
	_ = signal.Notify       // routes OS signals to a channel — https://pkg.go.dev/os/signal#Notify
	_ = slog.Info           // structured log line — https://pkg.go.dev/log/slog
	_ = errors.Is           // unwrap-aware error check — https://pkg.go.dev/errors#Is
	_ = time.Second         // time.Duration constant — https://pkg.go.dev/time#Second
	_ = os.Exit             // terminate the process with an exit code — https://pkg.go.dev/os#Exit
	_ = syscall.SIGINT      // signal sent by Ctrl-C — https://pkg.go.dev/syscall#SIGINT
	_ = http.NewServeMux    // creates a request multiplexer (router) — https://pkg.go.dev/net/http#NewServeMux
	_ = config.Load         // loads env vars into a Config struct
	_ = handler.Health      // GET /health handler func
	_ = handler.NewChat     // constructs the Chat handler
	_ = hf.NewClient        // constructs the HF API client
	_ = middleware.CORS     // CORS middleware wrapper
	_ = rag.Load            // loads SQuAD rows from HF Datasets Server
	_ = tools.NewExecutor   // constructs the tool executor
)

// main is the entry point — execution starts here.
//
// EXERCISE — implement this function.
//
// Concepts practiced:
//   - log/slog structured logging: slog.Info("msg", "key", val).
//     https://pkg.go.dev/log/slog
//   - := short variable declaration — declares and assigns in one step.
//     https://go.dev/ref/spec#Short_variable_declarations
//   - context.Background(): the root context; use when there is no parent.
//     https://pkg.go.dev/context#Background
//   - http.NewServeMux(): creates a new request multiplexer (router).
//     Go 1.22+ patterns support method prefixes: "GET /path", "POST /path".
//     https://pkg.go.dev/net/http#NewServeMux
//   - &http.Server{}: the server struct; set timeouts to prevent resource leaks.
//     ReadTimeout, WriteTimeout, IdleTimeout are the three important ones.
//     https://pkg.go.dev/net/http#Server
//   - os.Exit(1): terminates the process immediately with a non-zero exit code.
//     https://pkg.go.dev/os#Exit
//   - Goroutines: go func() { ... }() launches a concurrent function.
//     The trailing () immediately calls the anonymous function.
//     https://go.dev/tour/concurrency/1
//   - errors.Is(err, target): unwrap-aware comparison — works through wrapped errors.
//     https://pkg.go.dev/errors#Is
//   - Buffered channels: make(chan os.Signal, 1) — capacity 1 so signal.Notify
//     can send without a waiting receiver; prevents dropped signals.
//     https://go.dev/tour/concurrency/3
//   - signal.Notify(ch, sig...): routes OS signals to the channel.
//     https://pkg.go.dev/os/signal#Notify
//   - syscall.SIGINT / syscall.SIGTERM: Ctrl-C and kill/Docker-stop signals.
//     https://pkg.go.dev/syscall#SIGINT
//   - Blocking receive <-quit: pauses the goroutine until a value arrives on the channel.
//     https://go.dev/tour/concurrency/2
//   - context.WithTimeout for graceful shutdown: gives in-flight requests 10s to finish.
//     https://pkg.go.dev/context#WithTimeout
//   - defer cancel(): ensures the context's resources are freed even on early return.
//     https://go.dev/tour/flowcontrol/12
//
// Steps:
//
//  1. slog.Info("starting hf-go-chat server")
//
//  2. cfg, err := config.Load(). On error: slog.Error + os.Exit(1).
//
//  3. slog.Info("loading dataset", "name", cfg.DatasetName, "limit", cfg.DatasetLimit)
//     rows, err := rag.Load(context.Background(), cfg.DatasetName, cfg.DatasetCfg, cfg.DatasetLimit)
//     On error: slog.Error + os.Exit(1).
//     slog.Info("dataset loaded", "rows", len(rows))
//
//  4. Wire up dependencies:
//     hfClient  := hf.NewClient(cfg.HFToken, cfg.Model)
//     executor  := tools.NewExecutor(cfg.BraveKey)
//     chatHandler := handler.NewChat(hfClient, executor, rows)
//
//  5. Create the mux (http.NewServeMux) and register routes:
//     "GET /health"    → handler.Health (a plain func — use mux.HandleFunc)
//     "POST /api/chat" → chatHandler    (implements http.Handler — use mux.Handle)
//
//  6. Create the server:
//     srv := &http.Server{
//     Addr:         ":" + cfg.Port,
//     Handler:      middleware.CORS(mux),
//     ReadTimeout:  10 * time.Second,
//     WriteTimeout: 120 * time.Second,
//     IdleTimeout:  60 * time.Second,
//     }
//
//  7. Launch the server in a goroutine:
//     go func() {
//     slog.Info("listening", "addr", srv.Addr)
//     if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
//     slog.Error("server error", "err", err)
//     os.Exit(1)
//     }
//     }()
//
//  8. Set up graceful shutdown:
//     quit := make(chan os.Signal, 1)
//     signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
//     <-quit   // block until Ctrl-C or kill
//     slog.Info("shutting down")
//
//  9. Shutdown with a 10-second timeout:
//     shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//     defer cancel()
//     if err := srv.Shutdown(shutdownCtx); err != nil {
//     slog.Error("shutdown error", "err", err)
//     }
func main() {
	slog.Info("starting hf-go-chat server")
	cfg, err := config.Load()

	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Info("loading dataset", "name", cfg.DatasetName, "limit", cfg.DatasetLimit)

	rows, err := rag.Load(context.Background(), cfg.DatasetName, cfg.DatasetCfg, cfg.DatasetLimit)

	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	slog.Info("Dataset Loaded", "rows", len(rows))

	hfClient := hf.NewClient(cfg.HFToken, cfg.Model)
	executer := tools.NewExecutor(cfg.BraveKey)
	chatHandler := handler.NewChat(hfClient, executer, rows)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.Health)
	mux.Handle("POST /api/chat", chatHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      middleware.CORS(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting Down")

	shutdownCTX, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if err := srv.Shutdown(shutdownCTX); err != nil {
		slog.Error("shutdown error", "err", err)
	}

}
