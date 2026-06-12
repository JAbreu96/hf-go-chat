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

// Chat is a struct handler — it holds dependencies (hf client, executor, rag rows)
// instead of using global variables. This makes the handler testable and explicit.
type Chat struct {
	client   *hf.Client
	executor *tools.Executor
	rows     []rag.Row
}

// NewChat constructs a Chat handler.
func NewChat(client *hf.Client, executor *tools.Executor, rows []rag.Row) *Chat {
	return &Chat{
		client:   client,
		executor: executor,
		rows:     rows,
	}
}

// ServeHTTP implements http.Handler — a struct with a ServeHTTP method satisfies
// the interface automatically. We can register &Chat{} directly on the mux.
// https://pkg.go.dev/net/http#Handler
func (h *Chat) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// chatRequest is the JSON body sent by the Next.js client.
	var req struct {
		Messages []hf.Message `json:"messages"`
	}

	// json.NewDecoder(r.Body).Decode reads directly from the request body stream.
	// r.Body is an io.ReadCloser — it satisfies io.Reader so NewDecoder accepts it.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "messages array is required", http.StatusBadRequest)
		return
	}

	// RAG: find the last user message and retrieve relevant context from SQuAD.
	lastUserContent := lastUserMessage(req.Messages)
	contexts := rag.Retrieve(h.rows, lastUserContent, 3)

	systemPrompt := systemPromptBase
	if len(contexts) > 0 {
		systemPrompt += "\n\nReference material:\n" + strings.Join(contexts, "\n---\n")
	}

	// Prepend the system message. Slices grow with append; the system message
	// comes first so the model sees it before any conversation turns.
	messages := append([]hf.Message{{Role: "system", Content: systemPrompt}}, req.Messages...)

	// Set SSE headers before writing any body.
	// Once WriteHeader or Write is called the headers are sent and cannot be changed.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Type assertion with the comma-ok form: flusher, ok := w.(http.Flusher)
	// http.Flusher is an optional interface — not all ResponseWriters implement it.
	// The comma-ok form returns ok=false instead of panicking if it isn't implemented.
	// We need Flusher to push each SSE chunk to the client immediately.
	// https://pkg.go.dev/net/http#Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	toolDefs := []hf.Tool{tools.WebSearchTool()}

	slog.Info("chat request",
		"messages", len(req.Messages),
		"rag_contexts", len(contexts),
	)

	err := h.client.Run(
		r.Context(),
		messages,
		toolDefs,
		h.executor.Run,
		w,
		flusher.Flush,
	)
	if err != nil {
		// We may have already written SSE lines — can't change status code now.
		// Write an SSE error event so the client knows something went wrong.
		fmt.Fprintf(w, "data: {\"error\":%q}\n\n", err.Error())
		flusher.Flush()
		slog.Error("chat handler error", "err", err)
	}
}

// lastUserMessage finds the content of the most recent user message in the history.
func lastUserMessage(messages []hf.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			return messages[i].Content
		}
	}
	return ""
}
