// Package hf defines the request and response types for the Hugging Face
// Inference Providers chat completion API (OpenAI-compatible).
// API reference: https://huggingface.co/docs/inference-providers/tasks/chat-completion
package hf

// Message is a single turn in the conversation.
// Struct fields with a `json:"..."` tag control how they marshal to/from JSON.
// The tag name becomes the JSON key; omitempty skips the field when it is the zero value.
// https://pkg.go.dev/encoding/json#Marshal
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// Tool describes a function the model may call.
type Tool struct {
	Type     string       `json:"type"` // always "function"
	Function ToolFunction `json:"function"`
}

// ToolFunction is the schema of a callable tool.
type ToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// Parameters is raw JSON — we pass it straight through without parsing.
	// json.RawMessage delays decoding; useful when the shape varies.
	// https://pkg.go.dev/encoding/json#RawMessage
	Parameters interface{} `json:"parameters"`
}

// ToolCall is what the model returns when it wants to invoke a tool.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"` // always "function"
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction holds the name and JSON-encoded arguments the model chose.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // JSON string, needs a second json.Unmarshal
}

// ChatRequest is the body sent to /v1/chat/completions.
// Pointer fields (*bool, *int) let us distinguish "not set" from "set to zero/false".
// A nil pointer is omitted from JSON output via omitempty.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
	Stream   bool      `json:"stream"`
	// MaxTokens limits the response length. *int so we can omit it entirely when nil.
	MaxTokens *int `json:"max_tokens,omitempty"`
}

// ChatResponse is the non-streaming response body.
type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice is one candidate completion returned by the model.
type Choice struct {
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"` // "stop" | "tool_calls" | "length"
	Message      Message `json:"message"`
}

// Usage contains token consumption info.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
