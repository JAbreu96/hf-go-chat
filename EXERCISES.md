# Exercise 01 — Go Basics

Branch: `exercise/01-go-basics`

You are building the foundation of the app: the config loader and the RAG keyword retriever.
Every other file is complete — once you implement these two, the server will start and chat will work
(minus HTTP calls and streaming, which are exercise 02 and 03).

---

## Files to implement

### 1. `server/internal/config/config.go` → `Load()`

**What it does:** reads environment variables and returns a typed `Config` struct.

**Concepts you will use:**
- `os.Getenv` — reads a named env var, returns `""` if unset
- Multiple return values `(Config, error)` — Go's alternative to exceptions
- `if err != nil` — the canonical error check
- `errors.New` — creates a plain error value
- `strconv.Atoi` — parses a string as an integer
- Struct literals `Config{Field: value}`

**How to test:**
```bash
cd server
# Set env vars inline and run just the config package
HF_TOKEN=test BRAVE_API_KEY=test go run ./cmd/main.go
# Expected: server starts, logs "loading dataset..."
# (it will fail at the dataset fetch because HF_TOKEN is fake, that's fine)
```

---

### 2. `server/internal/rag/retriever.go` → `Retrieve()`

**What it does:** given a slice of SQuAD rows loaded in memory, scores each one against
the user's query by counting keyword matches, then returns the top-N context strings.

**Concepts you will use:**
- `strings.Fields` / `strings.ToLower` / `strings.Contains`
- Anonymous struct (inline struct type for scoring)
- `make([]T, 0, cap)` — pre-allocated slice
- `append`
- `sort.Slice` with a custom less function
- Blank identifier `_` in a range loop

**How to test:**
```bash
cd server
go build ./...   # must compile with no errors first
```
Then start the server with real tokens and ask a factual question — the server logs will show
how many RAG contexts were retrieved.

---

## Workflow

```bash
# Check your work compiles
cd server && go build ./...

# Run the server (needs a real .env)
go run ./cmd/main.go
```

When both functions are implemented and the server starts cleanly, move to:

```bash
git checkout exercise/02-go-http
```
