package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/joelchrist/hf-go-chat/internal/hf"
)

// Executor routes the model's tool_calls to the correct implementation.
// It is a struct so it can hold the API keys needed by each tool.
type Executor struct {
	braveKey string
}

// NewExecutor is the constructor for Executor.
func NewExecutor(braveKey string) *Executor {
	return &Executor{braveKey: braveKey}
}

// Run executes all tool calls the model requested and returns the combined results.
// The model may request multiple tools in one response; we run them in order.
func (e *Executor) Run(ctx context.Context, calls []hf.ToolCall) (string, error) {
	if len(calls) == 0 {
		return "", nil
	}

	// We only support one tool call per turn for simplicity.
	call := calls[0]

	switch call.Function.Name {
	case "web_search":
		return e.runWebSearch(ctx, call.Function.Arguments)
	default:
		return "", fmt.Errorf("unknown tool: %s", call.Function.Name)
	}
}

// runWebSearch parses the model's JSON arguments and calls Brave Search.
// The model encodes tool arguments as a JSON string inside the larger JSON response —
// so Arguments is a string like `{"query":"..."}` that needs a second json.Unmarshal.
func (e *Executor) runWebSearch(ctx context.Context, arguments string) (string, error) {
	// map[string]any is a generic JSON object: keys are strings, values are anything.
	// This is the idiomatic way to decode JSON when the shape is dynamic.
	// https://go.dev/tour/moretypes/19
	var args map[string]any
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("parsing web_search arguments: %w", err)
	}

	// Type assertion: the map value is `any` (interface{}); we assert it is a string.
	// The comma-ok form (v, ok := x.(T)) never panics — ok is false if the assertion fails.
	// https://go.dev/tour/methods/15
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("web_search requires a non-empty query string")
	}

	return Search(ctx, e.braveKey, query, 3)
}

// WebSearchTool returns the tool definition sent to the HF model.
// The model reads this to understand what tools are available and when to use them.
func WebSearchTool() hf.Tool {
	return hf.Tool{
		Type: "function",
		Function: hf.ToolFunction{
			Name:        "web_search",
			Description: "Search the web for current information not available in your training data. Use this for recent events, live data, or anything that may have changed since your knowledge cutoff.",
			// Parameters follows JSON Schema — the model uses this to know what
			// fields to include in its tool call arguments.
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
