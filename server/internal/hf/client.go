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

// ErrNoFlusher is a sentinel error value — a package-level variable of type error.
// Callers check: errors.Is(err, hf.ErrNoFlusher)
// https://pkg.go.dev/errors#New
var ErrNoFlusher = errors.New("response writer does not implement http.Flusher")

// Client holds credentials and the HTTP transport for all HF API calls.
// Unexported fields (lowercase) are only accessible within package hf.
// https://go.dev/ref/spec#Exported_identifiers
type Client struct {
	token      string
	model      string
	httpClient *http.Client
}

// Packages available for your implementation.
var (
	_ = bytes.NewReader   // wraps []byte as an io.Reader — https://pkg.go.dev/bytes#NewReader
	_ = bufio.NewScanner  // wraps io.Reader for line-by-line reading — https://pkg.go.dev/bufio#NewScanner
	_ = json.Marshal      // encodes a Go value to JSON bytes — https://pkg.go.dev/encoding/json#Marshal
	_ = json.NewDecoder   // wraps io.Reader for streaming JSON decode — https://pkg.go.dev/encoding/json#NewDecoder
	_ = fmt.Fprintf       // writes formatted string to any io.Writer — https://pkg.go.dev/fmt#Fprintf
	_ = io.ReadAll        // reads all bytes from an io.Reader — https://pkg.go.dev/io#ReadAll
	_ = time.Second       // time.Duration constant — https://pkg.go.dev/time#Second
)

// NewClient is a constructor — idiomatic Go uses NewX to return a pointer to a new value.
// Returning *Client means all callers share the same instance in memory.
// https://go.dev/doc/effective_go#composite_literals
//
// EXERCISE — implement this function.
//
// Steps:
//  1. Return &Client{ ... } with:
//       token: token
//       model: model
//       httpClient: &http.Client{Timeout: 90 * time.Second}
//     Use Go's zero values — only set Timeout; everything else defaults safely.
//     https://go.dev/ref/spec#The_zero_value
func NewClient(token, model string) *Client {
	panic("not implemented")
}

// Run is the agent loop. It calls HF once (non-streaming) to check for tool calls,
// executes any tools the model requested, then streams the final answer to w.
//
// Pointer receiver func (c *Client): c gives access to c.token, c.model without
// copying the whole struct. Use pointer receivers when the method reads struct state.
// https://go.dev/tour/methods/4
//
// context.Context is always the first parameter by convention — it carries
// deadlines and cancellation signals across API boundaries.
// https://pkg.go.dev/context
//
// EXERCISE — implement this function.
//
// Concepts practiced:
//   - for {} infinite loop with return to exit: Go's only loop keyword.
//     https://go.dev/tour/flowcontrol/1
//   - defer resp.Body.Close(): runs when the enclosing function returns.
//     Skipping this leaks TCP connections — Go's HTTP client reuses connections only
//     when the body is fully read and closed.
//     https://go.dev/tour/flowcontrol/12
//   - json.NewDecoder(resp.Body).Decode(&v): streams JSON from an io.Reader into a struct.
//     https://pkg.go.dev/encoding/json#Decoder.Decode
//   - errors.New: creates a plain error value from a string (no wrapping).
//     https://pkg.go.dev/errors#New
//   - io.ReadAll: reads all bytes from an io.Reader (used here to read error bodies).
//     https://pkg.go.dev/io#ReadAll
//   - append growing a slice mid-loop.
//     https://go.dev/tour/moretypes/15
//   - continue: skips to the next loop iteration.
//     https://go.dev/ref/spec#Continue_statements
//
// Steps:
//  1. Start an infinite for loop.
//  2. Call c.call(ctx, messages, tools, false) — non-streaming first call.
//     Return a wrapped error on failure.
//  3. defer resp.Body.Close() immediately after checking the error.
//  4. If resp.StatusCode != http.StatusOK, read body with io.ReadAll and return an error.
//  5. Decode the response into a ChatResponse. Return a wrapped error on failure.
//  6. If len(chat.Choices) == 0, return errors.New("hf returned no choices").
//  7. choice := chat.Choices[0]
//  8. If choice.FinishReason == "tool_calls":
//       a. Call executor(ctx, choice.Message.ToolCalls). Return a wrapped error on failure.
//       b. Append choice.Message to messages (the assistant's tool_call turn).
//       c. Append a Message{Role:"tool", Content:toolResult, ToolCallID: choice.Message.ToolCalls[0].ID}.
//       d. continue — loop back and call HF again with the tool result.
//  9. If finish_reason is anything else (typically "stop"): return c.stream(ctx, messages, w, flush).
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

// call makes one POST to the HF completions endpoint and returns the raw *http.Response.
// The caller is responsible for closing resp.Body.
//
// EXERCISE — implement this function.
//
// Concepts practiced:
//   - json.Marshal(v): encodes a Go value to JSON []byte.
//     https://pkg.go.dev/encoding/json#Marshal
//   - bytes.NewReader(b): wraps []byte as an io.Reader (required by http.NewRequestWithContext).
//     https://pkg.go.dev/bytes#NewReader
//   - http.NewRequestWithContext: creates an outbound request bound to a context.
//     https://pkg.go.dev/net/http#NewRequestWithContext
//   - req.Header.Set(key, value): sets a request header; call after NewRequestWithContext.
//     https://pkg.go.dev/net/http#Header.Set
//   - c.httpClient.Do(req): executes the request and returns (*http.Response, error).
//     https://pkg.go.dev/net/http#Client.Do
//
// Steps:
//  1. Build a ChatRequest{Model: c.model, Messages: messages, Tools: tools, Stream: stream}.
//  2. json.Marshal it into data []byte. Return a wrapped error on failure.
//  3. http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewReader(data)).
//     Return a wrapped error on failure.
//  4. Set headers: "Authorization" → "Bearer " + c.token, "Content-Type" → "application/json".
//  5. Return c.httpClient.Do(req).
func (c *Client) call(ctx context.Context, messages []Message, tools []Tool, stream bool) (*http.Response, error) {
	panic("not implemented")
}

// stream makes a streaming POST and writes each SSE line to w, flushing after each one.
//
// EXERCISE — implement this function.
//
// Concepts practiced:
//   - bufio.NewScanner(r): wraps any io.Reader; splits on newlines by default.
//     https://pkg.go.dev/bufio#Scanner
//   - scanner.Buffer(buf, max): raises the scanner's internal buffer limit.
//     Default is 64 KiB; SSE lines with large JSON payloads can exceed that.
//     https://pkg.go.dev/bufio#Scanner.Buffer
//   - scanner.Scan() / scanner.Text(): advance one line / return the current line.
//     https://pkg.go.dev/bufio#Scanner.Scan
//   - fmt.Fprintf(w, format, args): writes formatted output to any io.Writer.
//     http.ResponseWriter satisfies io.Writer because it has a Write([]byte) method.
//     https://pkg.go.dev/fmt#Fprintf
//   - flush(): calls http.Flusher.Flush — pushes buffered data to the client immediately.
//     https://pkg.go.dev/net/http#Flusher
//
// Steps:
//  1. Call c.call(ctx, messages, nil, true). Return a wrapped error on failure.
//     Defer resp.Body.Close().
//  2. If resp.StatusCode != http.StatusOK, read body and return a formatted error.
//  3. Create a bufio.NewScanner(resp.Body).
//     Call scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) to raise the limit.
//  4. Loop with scanner.Scan():
//       line := scanner.Text()
//       if line == "" { continue }
//       fmt.Fprintf(w, "%s\n\n", line)
//       flush()
//  5. Return scanner.Err().
func (c *Client) stream(ctx context.Context, messages []Message, w io.Writer, flush func()) error {
	panic("not implemented")
}
