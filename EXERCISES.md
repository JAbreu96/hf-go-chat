# Exercise: llm-concepts

## What to implement

All Go infrastructure is **complete** — HTTP server, config, RAG loading, Brave Search, middleware, streaming. You only implement the three pieces that are specific to LLM integration:

| File | Function | Concept |
|---|---|---|
| `server/internal/hf/client.go` | `Run()` | The agent loop — detect tool calls, execute, re-call |
| `server/internal/tools/executor.go` | `WebSearchTool()` | Tool definition schema (JSON Schema) |
| `server/internal/handler/chat.go` | Part of `ServeHTTP` | System prompt + RAG injection + tool list |

**Recommended order:** `WebSearchTool` → `ServeHTTP` (Parts A–E) → `Run`

---

## Concept 1 — Tool definition (`executor.go → WebSearchTool`)

The model cannot search the web on its own. You expose capabilities to it by sending
a **tool definition** alongside every chat request. The model reads the definition
and decides when and how to call it.

Tool definitions use **JSON Schema** to describe the arguments the model must provide.

```
hf.Tool {
  Type: "function"
  Function: {
    Name:        "web_search"
    Description: "..."   ← the model reads this to decide WHEN to use the tool
    Parameters: {        ← JSON Schema: tells the model WHAT fields to send
      "type": "object",
      "properties": { "query": { "type": "string", "description": "..." } },
      "required": ["query"]
    }
  }
}
```

References:
- https://huggingface.co/docs/inference-providers/tasks/chat-completion#tool-calling
- https://json-schema.org/understanding-json-schema/

---

## Concept 2 — System prompt + RAG injection (`handler/chat.go`)

LLMs have **no memory between requests** — you send the full conversation history
every time. To ground the model in your data, you:

1. Find the user's latest message
2. Search your in-memory SQuAD index for relevant passages (`rag.Retrieve`)
3. Build a **system message** — a special `role: "system"` turn that sets model
   behavior; it must be the **first** message in the array
4. Prepend it to the user's messages before sending to the API

```
messages = [
  { role: "system",    content: "You are helpful. Reference material:\n<squad passages>" },
  { role: "user",      content: "What is the capital of France?" },
]
```

Reference: https://huggingface.co/docs/inference-providers/tasks/chat-completion

---

## Concept 3 — The agent loop (`client.go → Run`)

When the model wants to use a tool, it returns `finish_reason: "tool_calls"` instead
of an answer. Your code must:

1. Detect the `"tool_calls"` finish reason
2. Execute the tool (Brave Search in this case)
3. Inject the result back into the message history as a `role: "tool"` message
4. Call the model **again** with the enriched history
5. Repeat until `finish_reason: "stop"` — then stream the final answer

```
for {
  resp = callHF(messages, stream=false)   // non-streaming to inspect finish_reason
  if resp.finish_reason == "tool_calls" {
    result = executeTool(resp.tool_calls)
    messages = append(messages, assistantTurn, toolResult)
    continue                               // loop — call HF again
  }
  stream(messages)                         // finish_reason == "stop" → stream answer
  return
}
```

Message roles in a tool-use turn:
```
user      → { role: "user",      content: "What happened today?" }
model     → { role: "assistant", tool_calls: [{name:"web_search", args:{query:"..."}}] }
you inject→ { role: "tool",      content: "<search results>", tool_call_id: "<id>" }
model     → { role: "assistant", content: "Here's what happened..." }  ← streamed
```

Reference: https://huggingface.co/docs/inference-providers/tasks/chat-completion#tool-calling

---

## How to test

```bash
# Must compile first
cd server && go build ./...

# Start the server (needs a real .env with HF_TOKEN + BRAVE_API_KEY)
go run ./cmd/main.go

# Test without tool use (RAG only)
curl -N -X POST http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"role":"user","content":"What is machine learning?"}]}'
# Expect: SSE stream of tokens

# Test with tool use (forces a Brave Search call)
curl -N -X POST http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"messages":[{"role":"user","content":"What happened in the news today?"}]}'
# Expect: server logs a Brave API call, then SSE stream with web-grounded answer
```

## Solution

See `main` branch for the complete reference implementation.
