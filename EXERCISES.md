# Exercise 03 — Go Streaming & Concurrency

Branch: `exercise/03-go-streaming`

The hardest layer. Everything from exercises 01 and 02 is already complete.
You will now implement the HF API client with its agent loop, the chat HTTP handler
with SSE streaming, and the `main` entry point with goroutines and graceful shutdown.

---

## Files to implement

### 1. `server/internal/hf/client.go` — `NewClient`, `Run`, `call`, `stream`

**What it does:** the core of the app. `NewClient` constructs the client;
`Run` is the agent loop — it calls HF, detects tool calls, executes them, and
re-calls until the model is done; `stream` pipes the final SSE response to
the HTTP writer.

**Concepts you will use:**
- Constructor: `&Client{field: value}`, `*http.Client{Timeout: ...}`
- Pointer receiver `(c *Client)` — method reads `c.token`, `c.model`
- `for {}` infinite loop with `continue` and `return` to control flow
- `defer resp.Body.Close()` inside the loop
- `json.NewDecoder(resp.Body).Decode(&chat)` — streaming JSON decode
- `append(messages, newMessage)` mid-loop to grow conversation history
- `bufio.NewScanner` + `scanner.Buffer` + `scanner.Scan()` / `scanner.Text()`
- `fmt.Fprintf(w, "%s\n\n", line)` — writing SSE frames to an `io.Writer`

---

### 2. `server/internal/handler/chat.go` — `ServeHTTP`

**What it does:** receives the POST request from Next.js, runs RAG retrieval,
prepends a system prompt, and hands off to the HF client for streaming.

**Concepts you will use:**
- Method validation (`r.Method != http.MethodPost`)
- `json.NewDecoder(r.Body).Decode` — `r.Body` is an `io.ReadCloser` satisfying `io.Reader`
- Building and prepending a system message with `append([]T{x}, rest...)`
- Setting SSE response headers before writing any body
- Type assertion with comma-ok: `flusher, ok := w.(http.Flusher)`
- `slog.Info` / `slog.Error` for structured logging
- Writing an SSE error frame after headers have already been sent

---

### 3. `server/cmd/main.go` — `main`

**What it does:** loads config, fetches the dataset, wires all dependencies, starts
the HTTP server in a goroutine, and blocks until a termination signal is received.

**Concepts you will use:**
- `context.Background()` — root context
- Goroutine: `go func() { ... }()` — server runs concurrently with main
- Buffered channel: `make(chan os.Signal, 1)` — capacity 1, never drops a signal
- `signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)` — route OS signals
- `<-quit` — blocking receive, pauses main goroutine until signal
- `context.WithTimeout` + `defer cancel()` for graceful shutdown
- `srv.Shutdown(shutdownCtx)` — drains in-flight requests cleanly

---

## Workflow

```bash
# Check compilation after each function
cd server && go build ./...

# Run the full server (needs a real .env)
go run ./cmd/main.go
```

**Test the agent loop (no tool use):**
```bash
curl -N -X POST http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"role":"user","content":"What is the capital of France?"}]}'
# Expect: SSE data: lines stream in, ends with data: [DONE]
```

**Test web search (forces tool use):**
```bash
curl -N -X POST http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"role":"user","content":"What happened in the news today?"}]}'
# Expect: server logs a Brave API call, then SSE stream with web-grounded answer
```

**Test graceful shutdown:**
```bash
go run ./cmd/main.go &
curl -s http://localhost:8080/health   # should return 200
kill -SIGTERM $!                       # server should log "shutting down" cleanly
```

---

## You're done! 🎉

If all three exercises pass, the full app is running. You've implemented:
- **Exercise 01:** constants, error handling, env vars, keyword retrieval
- **Exercise 02:** outbound HTTP, JSON decode, middleware, interfaces, type assertions
- **Exercise 03:** constructors, pointer receivers, streaming, goroutines, channels

The `main` branch has the complete reference implementation to compare against.
