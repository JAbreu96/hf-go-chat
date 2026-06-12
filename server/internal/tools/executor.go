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
// The model reads this schema to know the tool exists, what it does, and what
// arguments to provide when it decides to call it.
//
// EXERCISE — implement this function.
//
// ── Tool definitions and JSON Schema ─────────────────────────────────────────
// A tool definition has three parts:
//   Type     — always "function" for function-calling tools
//   Function — the actual definition (name, description, parameters)
//
// The "parameters" field uses JSON Schema to describe the expected arguments.
// JSON Schema reference: https://json-schema.org/understanding-json-schema/
//
// The model uses the description to decide WHEN to call the tool,
// and the parameters schema to know WHAT arguments to include.
// Tool calling reference:
//   https://huggingface.co/docs/inference-providers/tasks/chat-completion#tool-calling
//
// ── Steps ────────────────────────────────────────────────────────────────────
// Return an hf.Tool with:
//
//  Type: "function"
//
//  Function: hf.ToolFunction{
//    Name:        "web_search"
//    Description: "Search the web for current information not available in your
//                  training data. Use this for recent events, live data, or
//                  anything that may have changed since your knowledge cutoff."
//
//    Parameters: map[string]any{
//      "type": "object",
//      "properties": map[string]any{
//        "query": map[string]any{
//          "type":        "string",
//          "description": "The search query",
//        },
//      },
//      "required": []string{"query"},   // the model MUST provide this field
//    },
//  }
//
// The "required" field is critical — without it the model may omit the query
// argument and runWebSearch will return an error.
func WebSearchTool() hf.Tool {
	panic("not implemented")
}
