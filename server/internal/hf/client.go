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

// Run is the agent loop. It calls HF once (non-streaming) to check for tool calls,
// executes any requested tools, then streams the final answer to w.
//
// Pointer receiver func (c *Client): c is a pointer, so this method reads c.token
// etc. without copying the whole struct. Use pointer receivers when the method
// needs the struct's state or is large enough that copying would be wasteful.
// https://go.dev/tour/methods/4
//
// context.Context is always the first parameter by convention. It carries
// deadlines and cancellation signals across API call boundaries.
// https://pkg.go.dev/context
func (c *Client) Run(
	ctx context.Context,
	messages []Message,
	tools []Tool,
	executor func(ctx context.Context, calls []ToolCall) (string, error),
	w io.Writer,
	flush func(),
) error {
	// Agent loop — Go has one loop keyword: for.
	// for {} with no condition is an infinite loop; we exit with return.
	// https://go.dev/tour/flowcontrol/1
	for {
		// First call: stream=false so we receive a complete JSON response and
		// can inspect finish_reason before deciding whether to stream to the client.
		resp, err := c.call(ctx, messages, tools, false)
		if err != nil {
			return fmt.Errorf("hf call: %w", err)
		}
		// defer runs when the enclosing function returns, in LIFO order.
		// Always defer resp.Body.Close() — Go's HTTP client reuses TCP connections
		// only when the body is fully read and closed. Skipping this leaks connections.
		// https://pkg.go.dev/net/http#Response
		defer resp.Body.Close() //nolint:gocritic

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("hf API %d: %s", resp.StatusCode, string(body))
		}

		var chat ChatResponse
		// json.NewDecoder wraps the response body (an io.Reader) in a streaming
		// JSON decoder. Decode(&chat) fills in the struct fields from the JSON.
		if err := json.NewDecoder(resp.Body).Decode(&chat); err != nil {
			return fmt.Errorf("decoding hf response: %w", err)
		}

		if len(chat.Choices) == 0 {
			return errors.New("hf returned no choices")
		}

		choice := chat.Choices[0]

		// The model is requesting a tool call — execute it and loop.
		if choice.FinishReason == "tool_calls" {
			toolResult, err := executor(ctx, choice.Message.ToolCalls)
			if err != nil {
				return fmt.Errorf("tool execution: %w", err)
			}

			// append grows the slice by appending the element. If the backing
			// array has spare capacity it reuses it; otherwise it allocates a new
			// larger one. The slice header (pointer, len, cap) is updated.
			// https://go.dev/tour/moretypes/15
			messages = append(messages, choice.Message) // assistant turn with tool_calls
			messages = append(messages, Message{
				Role:       "tool",
				Content:    toolResult,
				ToolCallID: choice.Message.ToolCalls[0].ID,
			})
			continue // back to top — call HF again with the tool result injected
		}

		// finish_reason == "stop": no tool call, stream the final answer to the client.
		return c.stream(ctx, messages, w, flush)
	}
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
