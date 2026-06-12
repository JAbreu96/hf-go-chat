# hf-go-chat

A full-stack chat application built to learn Go, Next.js App Router, and LLM engineering simultaneously — three skills that come up constantly in job interviews.

## What it does

- **Chat** — streams responses token-by-token from a Hugging Face model
- **RAG** — loads 1 000 SQuAD Q&A rows at startup; retrieves the 3 most relevant contexts and injects them into every system prompt
- **Web search** — if the model decides it needs current information, it calls the `web_search` tool; the Go backend executes a Brave Search query and feeds the results back before streaming the final answer

## Architecture

```
Browser (Next.js)  ──POST /api/chat──►  Next.js Route Handler
                                                │ (server-side proxy)
                                                ▼
                                        Go server :8080
                                                │
                                    ┌───────────┼────────────┐
                                    ▼           ▼            ▼
                               RAG retrieval  HF API    Brave Search
                               (in-memory)   (tool      (tool execution)
                                             detection)
```

## Stack

| Layer | Tech |
|---|---|
| Backend language | Go (standard library only) |
| LLM | Hugging Face Inference Providers (`router.huggingface.co`) |
| Default model | `Qwen/Qwen2.5-7B-Instruct-1M` |
| RAG dataset | SQuAD via HF Datasets Server |
| Web search | Brave Search API |
| Frontend | Next.js 15 App Router + Tailwind CSS 3 |

## Prerequisites

- Go 1.22+
- Node.js 18.18+ (or 20+)
- [Hugging Face account](https://huggingface.co/settings/tokens) — create a token with `inference.serverless.write` permission
- [Brave Search API key](https://api.search.brave.com/app/keys) — free tier gives 2 000 req/month

## Setup

```bash
# 1. Go server
cd server
cp .env.example .env
# Edit .env — set HF_TOKEN and BRAVE_API_KEY

# 2. Next.js client
cd ../client
cp .env.local.example .env.local
# GO_BACKEND_URL is already set to http://localhost:8080
npm install
```

## Run

Two terminals:

```bash
# Terminal 1 — Go server
cd server
go run ./cmd/main.go
# Logs: dataset loading (takes a few seconds), then "listening :8080"

# Terminal 2 — Next.js client
cd client
npm run dev
# Open http://localhost:3000
```

## Smoke tests

**No tool use (answered from RAG context):**
```bash
curl -N -X POST http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"role":"user","content":"What is the capital of France?"}]}'
```
Expect: SSE `data:` lines stream in, end with `data: [DONE]`.

**Forces web search:**
```bash
curl -N -X POST http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"role":"user","content":"What happened in the news today?"}]}'
```
Expect: server logs show a Brave API call, then the answer streams with web-sourced content.

## Go concepts covered

Every concept below appears organically in the source — comments in each file name the concept and link to the official doc.

| Group | Concepts |
|---|---|
| Language basics | `package`, `import`, exported vs unexported (capitalization), `:=` vs `var`, blank identifier `_`, constants |
| Functions | Multiple return values `(T, error)`, anonymous functions, closures, first-class function values, variadic functions |
| Types | Structs + JSON tags, pointer fields (`*T`, `omitempty`), slices, maps, zero values, pointers |
| OOP | Struct methods, pointer receivers, constructor pattern `NewX() *X`, implicit interface satisfaction |
| Interfaces | `io.Reader`, `io.ReadCloser`, `http.Handler`, `http.HandlerFunc`, `http.Flusher`, type assertions (comma-ok form) |
| Error handling | `if err != nil`, sentinel errors, `fmt.Errorf` with `%w` wrapping, `errors.Is` |
| Defer | `defer resp.Body.Close()` — resource cleanup, LIFO order |
| Concurrency | Goroutines (`go func()`), buffered channels, blocking receive `<-ch`, `context.WithTimeout` |
| Agent loop | `for {}` infinite loop, `append` mid-loop, two-pass JSON unmarshal (tool call arguments), `url.QueryEscape` |
| Standard library | `net/http`, `encoding/json`, `bufio`, `strings`, `sort`, `net/url`, `os`, `context`, `log/slog`, `fmt`, `io`, `errors`, `syscall` |
| Go modules | `go mod init`, `go.mod`, `go mod tidy`, `go.sum` |

## Next steps

- Add Go tests (`testing` package — `go test ./...`)
- Persist conversation history in a database (learn `database/sql`)
- Add auth (learn middleware chaining)
- Deploy the Go server as a Docker container
