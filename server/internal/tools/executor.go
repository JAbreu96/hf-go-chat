package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joelchrist/hf-go-chat/internal/hf"
)

// Executor routes the model's tool_calls to the correct implementation.
type Executor struct {
	braveKey string
}

// NewExecutor is the constructor for Executor.
func NewExecutor(braveKey string) *Executor {
	return &Executor{braveKey: braveKey}
}

// Packages available for your implementation.
var (
	_ = json.Unmarshal // decodes JSON bytes into a Go value
	_ = fmt.Errorf     // creates a formatted error, supports %w for wrapping
)

// Run executes all tool calls the model requested and returns the combined result.
//
// EXERCISE — implement this function.
//
// Concepts practiced:
//   - Switch statement on a string: routes different tool names to handlers.
//     https://go.dev/tour/flowcontrol/9
//   - Returning a formatted error for the default case (unknown tool name).
//     fmt.Errorf("unknown tool: %s", call.Function.Name)
//
// Steps:
//  1. If len(calls) == 0, return "", nil immediately.
//  2. Take only the first call: call := calls[0].
//  3. Switch on call.Function.Name:
//       case "web_search": return e.runWebSearch(ctx, call.Function.Arguments)
//       default: return "", fmt.Errorf("unknown tool: %s", call.Function.Name)
func (e *Executor) Run(ctx context.Context, calls []hf.ToolCall) (string, error) {
	panic("not implemented")
}

// runWebSearch parses the model's JSON-encoded arguments and calls Brave Search.
//
// EXERCISE — implement this function.
//
// Concepts practiced:
//   - Two-pass JSON unmarshal: the model's tool arguments arrive as a JSON string
//     (e.g. `{"query":"..."}`) embedded inside the outer JSON response. We need a
//     second json.Unmarshal call to decode that inner string.
//   - map[string]any: the idiomatic Go type for a JSON object with unknown keys.
//     https://go.dev/tour/moretypes/19
//   - Type assertion with comma-ok: v, ok := x.(T)
//     Never use bare x.(T) on an interface — it panics if T is wrong.
//     Use the comma-ok form so you can return a clean error instead.
//     https://go.dev/tour/methods/15
//
// Steps:
//  1. Declare: var args map[string]any
//     Call json.Unmarshal([]byte(arguments), &args).
//     Return a wrapped error on failure.
//  2. Extract the query with a type assertion:
//       query, ok := args["query"].(string)
//     If !ok or query == "", return an error: "web_search requires a non-empty query string"
//  3. Call Search(ctx, e.braveKey, query, 3) and return its result.
func (e *Executor) runWebSearch(ctx context.Context, arguments string) (string, error) {
	panic("not implemented")
}

// WebSearchTool returns the tool definition sent to the HF model.
// This is complete — no exercise here. Read it to understand how tool schemas work.
func WebSearchTool() hf.Tool {
	return hf.Tool{
		Type: "function",
		Function: hf.ToolFunction{
			Name:        "web_search",
			Description: "Search the web for current information not available in your training data. Use this for recent events, live data, or anything that may have changed since your knowledge cutoff.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query",
					},
				},
				"required": []string{"query"},
			},
		},
	}
}
