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
	_ = json.NewDecoder // decode request body — https://pkg.go.dev/encoding/json#NewDecoder
	_ = http.Error      // write an error status + plain-text body — https://pkg.go.dev/net/http#Error
	_ = strings.Join    // join a []string with a separator — https://pkg.go.dev/strings#Join
	_ = fmt.Fprintf     // write formatted string to any io.Writer — https://pkg.go.dev/fmt#Fprintf
	_ = slog.Info       // structured log line — https://pkg.go.dev/log/slog
)

// ServeHTTP implements http.Handler. A struct with a ServeHTTP method satisfies
// the http.Handler interface automatically — no `implements` keyword needed.
// https://pkg.go.dev/net/http#Handler
//
// EXERCISE — implement this method.
//
// Concepts practiced:
//   - Method validation: reject non-POST requests.
//     http.Error(w, msg, code) writes the status code and a plain-text body in one call.
//     https://pkg.go.dev/net/http#Error
//   - json.NewDecoder(r.Body).Decode(&v): r.Body is an io.ReadCloser; NewDecoder accepts io.Reader.
//     https://pkg.go.dev/encoding/json#Decoder.Decode
//   - strings.Join(slice, sep): concatenates a []string with a separator between elements.
//     https://pkg.go.dev/strings#Join
//   - SSE response headers: Content-Type "text/event-stream", Cache-Control "no-cache",
//     Connection "keep-alive" — must be set before any body bytes are written.
//     w.Header().Set(key, value) — sets a single header. https://pkg.go.dev/net/http#Header.Set
//     Reference: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
//   - http.Flusher type assertion (comma-ok form):
//     flusher, ok := w.(http.Flusher)
//     Not all ResponseWriters implement Flusher — always use the comma-ok form to avoid a panic.
//     https://pkg.go.dev/net/http#Flusher
//   - slog.Info("event", "key", value): structured key-value logging.
//     https://pkg.go.dev/log/slog
//   - fmt.Fprintf(w, format, args): writes a formatted string to any io.Writer.
//     Used here to send an SSE error frame after headers are already sent.
//     https://pkg.go.dev/fmt#Fprintf
//
// Steps:
//
//  1. If r.Method != http.MethodPost, call http.Error(w, "method not allowed", 405) and return.
//
//  2. Decode the request body into:
//     var req struct { Messages []hf.Message `json:"messages"` }
//     On error: http.Error(w, "invalid JSON body", http.StatusBadRequest) and return.
//     If len(req.Messages) == 0: http.Error(w, "messages array is required", 400) and return.
//
//  3. RAG retrieval:
//     lastUserContent := lastUserMessage(req.Messages)
//     contexts := rag.Retrieve(h.rows, lastUserContent, 3)
//     Build systemPrompt: start with systemPromptBase; if len(contexts) > 0, append:
//     "\n\nReference material:\n" + strings.Join(contexts, "\n---\n")
//
//  4. Prepend system message:
//     messages := append([]hf.Message{{Role:"system", Content:systemPrompt}}, req.Messages...)
//
//  5. Set SSE response headers on w (three calls to w.Header().Set).
//
//  6. Type-assert w to http.Flusher (comma-ok). If !ok, http.Error 500 and return.
//
//  7. slog.Info("chat request", "messages", len(req.Messages), "rag_contexts", len(contexts))
//
//  8. Call h.client.Run(r.Context(), messages, []hf.Tool{tools.WebSearchTool()},
//     h.executor.Run, w, flusher.Flush)
//     If err != nil:
//     fmt.Fprintf(w, "data: {\"error\":%q}\n\n", err.Error())
//     flusher.Flush()
//     slog.Error("chat handler error", "err", err)
func (h *Chat) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}

	var mes struct {
		Messages []hf.Message `json:"messages"`
	}

	if err := json.NewDecoder(r.Body).Decode(&mes); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if len(mes.Messages) == 0 {
		http.Error(w, "messages array is required", 400)
		return
	}

	lastUserContent := lastUserMessage(mes.Messages)

	contexts := rag.Retrieve(h.rows, lastUserContent, 3)

	systemPrompt := make([]string, 0, 1000)
	systemPrompt = append(systemPrompt, systemPromptBase)

	if len(contexts) > 0 {
		systemPrompt = append(systemPrompt, "\n\nReference Material:\n", strings.Join(contexts, "\n---\n"))
	}

	messages := append([]hf.Message{{Role: "system", Content: strings.Join(systemPrompt, "\n---\n")}}, mes.Messages...)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)

	if !ok {
		http.Error(w, "Response Writer is not of type http.Flusher", 500)
		return
	}

	slog.Info("Chat Request", "Messages", len(mes.Messages), "Rag Contexts", len(contexts))

	err := h.client.Run(r.Context(), messages, []hf.Tool{tools.WebSearchTool()}, h.executor.Run, w, flusher.Flush)

	if err != nil {
		fmt.Fprintf(w, "data: {\"error\":%q}\n\n", err.Error())
		flusher.Flush()
		slog.Error("chat handler error", "err", err)
	}

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
