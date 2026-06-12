# Exercise 02 — Go HTTP & Interfaces

Branch: `exercise/02-go-http`

You've already implemented the basics (config + retriever are complete from `main`).
Now you'll build the HTTP layer: making outbound requests, reading JSON responses,
and learning Go's interface-based middleware pattern.

---

## Files to implement

### 1. `server/internal/rag/loader.go` → `Load()`

**What it does:** calls the Hugging Face Datasets Server API at startup, decodes the
JSON response, and returns a flat slice of SQuAD rows stored in memory.

**Concepts you will use:**
- `url.Values` — query string builder; `.Encode()` percent-encodes values
- `http.NewRequestWithContext` — outbound request bound to a context
- `defer resp.Body.Close()` — resource cleanup (first real encounter with defer)
- `json.NewDecoder(r).Decode(&v)` — streaming JSON decode
- `fmt.Errorf("context: %w", err)` — error wrapping with `%w`
- `make([]Row, 0, n)` + `append`

---

### 2. `server/internal/tools/brave.go` → `Search()`

**What it does:** calls Brave Search with the model's query and formats the top
results as a string the model can read.

**Concepts you will use:**
- Same HTTP + JSON pattern as `loader.go`, but with custom request headers
- `req.Header.Set` — adding auth and accept headers
- `strings.Builder` + `fmt.Fprintf(&sb, ...)` — building a string incrementally

---

### 3. `server/internal/middleware/cors.go` → `CORS()`

**What it does:** wraps any `http.Handler` with CORS response headers so the
Next.js dev server can call Go without browser preflight errors.

**Concepts you will use:**
- `http.Handler` interface — one method: `ServeHTTP(ResponseWriter, *Request)`
- `http.HandlerFunc` — a named function type that implements `http.Handler`
- The middleware pattern: `func(next http.Handler) http.Handler`
- `w.Header().Set`, `w.WriteHeader`, `next.ServeHTTP`

---

### 4. `server/internal/tools/executor.go` → `Run()` and `runWebSearch()`

**What it does:** receives the model's `tool_calls`, dispatches to the right
implementation, and returns the result string.

**Concepts you will use:**
- `switch` on a string value
- `map[string]any` — generic JSON object type
- Two-pass JSON unmarshal: the tool arguments are a JSON *string* that needs
  a second `json.Unmarshal` call
- Type assertion `v, ok := x.(string)` — comma-ok form, never panics

---

## Workflow

```bash
# Check compilation after each file
cd server && go build ./...

# Run the server (needs a real .env with HF_TOKEN + BRAVE_API_KEY)
go run ./cmd/main.go
```

When all four functions compile and the server starts, test CORS:
```bash
curl -X OPTIONS http://localhost:8080/api/chat \
  -H "Origin: http://localhost:3001" -v 2>&1 | grep "Access-Control"
# Should print the Access-Control-Allow-* headers
```

When done, move to the final exercise:

```bash
git checkout exercise/03-go-streaming
```
