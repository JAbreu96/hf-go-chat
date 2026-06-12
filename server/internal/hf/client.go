package hf

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://router.huggingface.co/v1/chat/completions"

// ErrNoFlusher is a sentinel error — a package-level variable of type error.
// Callers check for it with errors.Is(err, hf.ErrNoFlusher).
// https://pkg.go.dev/errors#New
var ErrNoFlusher = errors.New("response writer does not implement http.Flusher")

// Client holds the credentials and HTTP transport for all HF API calls.
// Fields are unexported (lowercase) — only code inside package hf can access them.
// This is Go's encapsulation: no private/protected keywords, just case.
// https://go.dev/ref/spec#Exported_identifiers
type Client struct {
	token      string
	model      string
	httpClient *http.Client
}

// NewClient is a constructor — idiomatic Go uses a plain function named NewX
// that returns a pointer to the newly allocated value.
// Returning *Client means all callers share the same Client in memory.
// https://go.dev/doc/effective_go#composite_literals
func NewClient(token, model string) *Client {
	return &Client{
		token: token,
		model: model,
		// http.Client{} uses Go's zero values for unset fields.
		// We only override Timeout; everything else defaults safely.
		// https://go.dev/ref/spec#The_zero_value
		httpClient: &http.Client{Timeout: 90 * time.Second},
	}
}

// Run is the agent loop — the core of LLM-powered agentic behavior.
// NewClient, call, and stream are complete; this is the only function to implement.
//
// EXERCISE — implement this function.
//
// ── How the HF chat completions API works ────────────────────────────────────
// Every request sends a full array of messages and receives one response.
// The response has a "finish_reason" field that tells you why the model stopped:
//
//   "stop"       — normal completion; content holds the final answer
//   "tool_calls" — the model wants to call a function; no answer yet
//
// API reference: https://huggingface.co/docs/inference-providers/tasks/chat-completion
//
// ── Why we call non-streaming first ──────────────────────────────────────────
// Streaming delivers tokens one-by-one — we can't inspect finish_reason until
// all tokens arrive. A non-streaming call returns complete JSON immediately,
// so we can branch on finish_reason before committing to a streaming response.
//
// ── Message roles ─────────────────────────────────────────────────────────────
// The conversation is a flat array; roles tell the model who said each line:
//   "system"    — invisible instructions prepended before the conversation
//   "user"      — a message from the human
//   "assistant" — a model turn (may carry ToolCalls instead of Content)
//   "tool"      — the output of a tool, addressed back to the model
//
// A tool-use turn sequence looks like:
//   user → assistant(tool_calls) → tool(result) → assistant(final answer, streamed)
//
// Tool calling reference:
//   https://huggingface.co/docs/inference-providers/tasks/chat-completion#tool-calling
//
// ── Steps ────────────────────────────────────────────────────────────────────
//  1. Start an infinite for loop — for {} in Go.
//     https://go.dev/tour/flowcontrol/1
//
//  2. Call c.call(ctx, messages, tools, false) — non-streaming.
//     Return a wrapped error on failure (fmt.Errorf("hf call: %w", err)).
//     Immediately defer resp.Body.Close() — required to reuse TCP connections.
//     https://pkg.go.dev/net/http#Response
//
//  3. If resp.StatusCode != http.StatusOK, read body with io.ReadAll and return
//     fmt.Errorf("hf API %d: %s", resp.StatusCode, string(body)).
//
//  4. Decode the body: var chat ChatResponse
//     json.NewDecoder(resp.Body).Decode(&chat)
//     https://pkg.go.dev/encoding/json#Decoder.Decode
//
//  5. Guard: if len(chat.Choices) == 0 { return errors.New("hf returned no choices") }
//
//  6. choice := chat.Choices[0]
//
//  7. If choice.FinishReason == "tool_calls":
//       a. Call executor(ctx, choice.Message.ToolCalls) — runs the real tool.
//       b. Append choice.Message to messages.
//          (The assistant's tool_call turn must stay in history so the model
//           knows what it requested.)
//       c. Append Message{Role: "tool", Content: toolResult,
//            ToolCallID: choice.Message.ToolCalls[0].ID}
//          ToolCallID links this result to the specific call the model made.
//       d. continue — loop back to step 2 with the enriched message history.
//
//  8. finish_reason is "stop" — the model is done with tool use.
//     Return c.stream(ctx, messages, w, flush) to stream the final answer.
func (c *Client) Run(
	ctx context.Context,
	messages []Message,
	tools []Tool,
	executor func(ctx context.Context, calls []ToolCall) (string, error),
	w io.Writer,
	flush func(),
) error {
	panic("not implemented")
}

// call makes one POST to the HF completions endpoint and returns the raw response.
func (c *Client) call(ctx context.Context, messages []Message, tools []Tool, stream bool) (*http.Response, error) {
	body := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
		Stream:   stream,
	}

	// json.Marshal encodes body as JSON bytes.
	// bytes.NewReader wraps []byte as an io.Reader — the type http.NewRequestWithContext expects.
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	// http.NewRequestWithContext attaches ctx so the outbound request is cancelled
	// if the handler's context is cancelled (e.g. the browser disconnects).
	// https://pkg.go.dev/net/http#NewRequestWithContext
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

// stream makes a streaming POST and writes each SSE line to w, flushing after each.
func (c *Client) stream(ctx context.Context, messages []Message, w io.Writer, flush func()) error {
	resp, err := c.call(ctx, messages, nil, true)
	if err != nil {
		return fmt.Errorf("hf stream call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("hf stream API %d: %s", resp.StatusCode, string(body))
	}

	// bufio.NewScanner wraps any io.Reader and splits on newlines by default.
	// It reads the SSE response body one line at a time without loading it all into memory.
	// https://pkg.go.dev/bufio#Scanner
	scanner := bufio.NewScanner(resp.Body)
	// Increase the scanner buffer to handle large SSE data lines.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		// fmt.Fprintf writes a formatted string to any io.Writer.
		// w is an http.ResponseWriter — it satisfies io.Writer because it has Write([]byte).
		// https://pkg.go.dev/fmt#Fprintf
		fmt.Fprintf(w, "%s\n\n", line)
		flush()
	}
	return scanner.Err()
}
