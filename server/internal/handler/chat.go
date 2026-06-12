package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/joelchrist/hf-go-chat/internal/hf"
	"github.com/joelchrist/hf-go-chat/internal/rag"
	"github.com/joelchrist/hf-go-chat/internal/tools"
)

const systemPromptBase = `You are a helpful assistant with access to a knowledge base and the web.
When answering, prefer information from the provided reference material if it is relevant.
If you need current or real-time information, use the web_search tool.`

// Chat is a struct handler — it holds dependencies instead of using global variables.
// This makes the handler testable and keeps state explicit.
type Chat struct {
	client   *hf.Client
	executor *tools.Executor
	rows     []rag.Row
}

// NewChat constructs a Chat handler with its dependencies.
func NewChat(client *hf.Client, executor *tools.Executor, rows []rag.Row) *Chat {
	return &Chat{client: client, executor: executor, rows: rows}
}

// Packages available for your implementation.
var (
	_ = json.NewDecoder    // decode request body
	_ = http.Error         // write an error status + plain-text body
	_ = strings.Join       // join a []string with a separator
	_ = fmt.Fprintf        // write formatted string to any io.Writer
	_ = slog.Info          // structured log line
)

// ServeHTTP implements http.Handler. A struct with a ServeHTTP method satisfies
// the http.Handler interface automatically — no `implements` keyword needed.
// https://pkg.go.dev/net/http#Handler
//
// EXERCISE — implement this method.
//
// Concepts practiced:
//   - Method validation: reject non-POST requests with http.Error + http.StatusMethodNotAllowed.
//   - json.NewDecoder(r.Body).Decode(&v): r.Body is an io.ReadCloser; NewDecoder accepts io.Reader.
//   - SSE response headers: Content-Type "text/event-stream", Cache-Control "no-cache",
//     Connection "keep-alive" — must be set BEFORE any body is written.
//   - http.Flusher type assertion (comma-ok form):
//       flusher, ok := w.(http.Flusher)
//     Not all ResponseWriters implement Flusher — always use the comma-ok form.
//     https://pkg.go.dev/net/http#Flusher
//   - slog.Info("event", "key", value): structured logging.
//     https://pkg.go.dev/log/slog
//   - fmt.Fprintf(w, "data: {\"error\":%q}\n\n", err): writing an SSE error frame
//     after headers have already been sent (can't change status code at this point).
//
// Steps:
//  1. If r.Method != http.MethodPost, call http.Error(w, "method not allowed", 405) and return.
//
//  2. Decode the request body into:
//       var req struct { Messages []hf.Message `json:"messages"` }
//     On error: http.Error(w, "invalid JSON body", http.StatusBadRequest) and return.
//     If len(req.Messages) == 0: http.Error(w, "messages array is required", 400) and return.
//
//  3. RAG retrieval:
//       lastUserContent := lastUserMessage(req.Messages)
//       contexts := rag.Retrieve(h.rows, lastUserContent, 3)
//     Build systemPrompt: start with systemPromptBase; if len(contexts) > 0, append:
//       "\n\nReference material:\n" + strings.Join(contexts, "\n---\n")
//
//  4. Prepend system message:
//       messages := append([]hf.Message{{Role:"system", Content:systemPrompt}}, req.Messages...)
//
//  5. Set SSE response headers on w (three calls to w.Header().Set).
//
//  6. Type-assert w to http.Flusher (comma-ok). If !ok, http.Error 500 and return.
//
//  7. slog.Info("chat request", "messages", len(req.Messages), "rag_contexts", len(contexts))
//
//  8. Call h.client.Run(r.Context(), messages, []hf.Tool{tools.WebSearchTool()},
//       h.executor.Run, w, flusher.Flush)
//     If err != nil:
//       fmt.Fprintf(w, "data: {\"error\":%q}\n\n", err.Error())
//       flusher.Flush()
//       slog.Error("chat handler error", "err", err)
func (h *Chat) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	panic("not implemented")
}

// lastUserMessage finds the content of the most recent user turn.
// This is complete — no exercise here.
func lastUserMessage(messages []hf.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	return ""
}
