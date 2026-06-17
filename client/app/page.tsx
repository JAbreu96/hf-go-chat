"use client";
// 'use client' marks this module as a Client Component — it runs in the browser
// and can use React hooks (useState, useRef, useEffect). Without this directive,
// Next.js treats every component in app/ as a Server Component by default.
// https://nextjs.org/docs/app/building-your-application/rendering/client-components

import { useEffect, useRef, useState } from "react";

// --- Types ---

type Role = "user" | "assistant";

interface Message {
  id: string;
  role: Role;
  content: string;
}

// --- SSE parsing ---

// parseDelta extracts the text token from one SSE data line.
// The HF streaming format mirrors OpenAI:
//   data: {"choices":[{"delta":{"content":"token"},"finish_reason":null}]}
// The final line is: data: [DONE]
// https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
function parseDelta(line: string): string | null {
  if (!line.startsWith("data: ")) return null;
  const payload = line.slice(6).trim();
  if (payload === "[DONE]") return null;
  try {
    const json = JSON.parse(payload);
    return json?.choices?.[0]?.delta?.content ?? null;
  } catch {
    return null;
  }
}

// --- Component ---

export default function ChatPage() {
  // useState<T>(initial) declares a state variable and a setter.
  // The generic T pins the type; TypeScript infers it from the initial value when possible.
  // https://react.dev/reference/react/useState
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [isStreaming, setIsStreaming] = useState(false);

  // useRef holds a mutable value that does NOT trigger re-renders when changed.
  // Here we store a DOM ref to the scroll anchor div at the bottom of the message list.
  // https://react.dev/reference/react/useRef
  const bottomRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to the latest message after every render where messages changed.
  // useEffect(fn, [dep]) runs fn after the DOM updates whenever dep changes.
  // https://react.dev/reference/react/useEffect
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  // EXERCISE: implement handleSubmit.
  //
  // This function fires when the user submits the chat form. It should:
  //   1. Prevent the default form submission (full page reload).
  //   2. Optimistically add the user message and an empty assistant message to state.
  //   3. POST the conversation to /api/chat (the Next.js Route Handler proxy).
  //   4. Read the streaming SSE response body chunk-by-chunk.
  //   5. Decode each chunk, parse the delta token, and append it to the assistant message.
  //
  // Step 1 — Guard against empty input or double submission.
  //   e.preventDefault()
  //   const text = input.trim(); if (!text || isStreaming) return;
  //
  // Step 2 — Optimistically add both messages and reset form state.
  //   const userMsg: Message = { id: crypto.randomUUID(), role: "user", content: text }
  //   const assistantMsg: Message = { id: crypto.randomUUID(), role: "assistant", content: "" }
  //   setMessages((prev) => [...prev, userMsg, assistantMsg])
  //   setInput(""); setIsStreaming(true)
  //   https://react.dev/reference/react/useState#updating-state-based-on-the-previous-state
  //
  // Step 3 — POST to the Next.js Route Handler.
  //   fetch("/api/chat", { method: "POST", headers: { "Content-Type": "application/json" },
  //     body: JSON.stringify({ messages: [...messages, userMsg].map(({ role, content }) => ({ role, content })) }) })
  //   Throw if !res.ok or !res.body.
  //   https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API/Using_Fetch
  //
  // Step 4 — Obtain a reader and decoder to consume the stream incrementally.
  //   const reader = res.body.getReader()
  //   const decoder = new TextDecoder()
  //   let buffer = ""
  //   Loop: const { value, done } = await reader.read(); if (done) break;
  //   https://developer.mozilla.org/en-US/docs/Web/API/ReadableStream/getReader
  //
  // Step 5 — Decode the Uint8Array chunk, keeping multi-byte char state intact.
  //   buffer += decoder.decode(value, { stream: true })
  //   { stream: true } prevents flushing the decoder's internal state between calls —
  //   necessary when a UTF-8 character (e.g. emoji) spans two chunks.
  //   https://developer.mozilla.org/en-US/docs/Web/API/TextDecoder/decode
  //
  // Step 6 — Split the buffer on "\n" and parse each complete SSE line.
  //   const lines = buffer.split("\n"); buffer = lines.pop() ?? ""
  //   The last element may be an incomplete line — keep it in buffer for the next chunk.
  //   Call parseDelta(line.trim()) on each line; it returns the token string or null.
  //
  // Step 7 — Append the token to the assistant message with a functional setState.
  //   setMessages((prev) =>
  //     prev.map((m) => m.id === assistantMsg.id ? { ...m, content: m.content + delta } : m)
  //   )
  //   WHY functional update: the callback always receives the latest state snapshot,
  //   preventing stale closures from overwriting tokens already streamed in.
  //   https://react.dev/reference/react/useState#updating-state-based-on-the-previous-state
  //
  // Step 8 — Handle errors and always clear isStreaming.
  //   catch: setMessages with the error string as the assistant content
  //   finally: setIsStreaming(false)
  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const text = input.trim();

    if (!text || isStreaming) {
      return;
    }

    const userMsg: Message = { id: crypto.randomUUID(), role: 'user', content: text }
    const agentMsg: Message = { id: crypto.randomUUID(), role: 'assistant', content: "" }

    setMessages((prev) => [...prev, userMsg, agentMsg])
    setInput("")
    setIsStreaming(true)

    const REQ_BODY = {
      messages: [...messages, userMsg].map(({ role, content }) => ({ role, content }))
    }

    const init: RequestInit = {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(REQ_BODY)
    }

    try {
      const res = await fetch("/api/chat", init)
      if (!res.ok || !res.body) {
        throw new Error("Message failed to send. Try again later")
      }
      await readStream(res, agentMsg);
    } catch (e) {
      setMessages((prev) =>
        prev.map((m) => m.id === agentMsg.id ? { ...m, content: "" + e } : m)
      );
    } finally {
      setIsStreaming(false);
    }
  }

  async function readStream(res: Response, agentMsg: Message) {
    const reader = res.body?.getReader()
    const decoder = new TextDecoder()

    let buffer = ""

    while (true && reader) {
      const { value, done } = await reader.read();

      if (done) {
        break
      }
      buffer += decoder.decode(value, { stream: true })

      const lines = buffer.split("\n")

      buffer = lines.pop() ?? ""

      for (let i = 0; i < lines.length; i++) {
        const line = lines[i];

        const token = parseDelta(line.trim())

        setMessages(prev => prev.map((message) => {
          if (message.id === agentMsg.id) {
            return { ...message, content: message.content + `${token ?? ""}` }
          } else {
            return message
          }
        }))
      }
    }
  }

  return (
    <div className="flex flex-col h-screen bg-gray-950 text-gray-100 font-mono">
      {/* Header */}
      <header className="border-b border-gray-800 px-6 py-4 shrink-0">
        <h1 className="text-lg font-semibold tracking-tight">HF Go Chat</h1>
        <p className="text-xs text-gray-500 mt-0.5">
          RAG + web search · Hugging Face · Go backend
        </p>
      </header>

      {/* Message list */}
      <main className="flex-1 overflow-y-auto px-4 py-6 space-y-4">
        {messages.length === 0 && (
          <p className="text-center text-gray-600 text-sm mt-16">
            Ask anything — the model has a SQuAD knowledge base and can search the web.
          </p>
        )}
        {messages.map((msg) => (
          <div
            key={msg.id}
            className={`flex ${msg.role === "user" ? "justify-end" : "justify-start"}`}
          >
            <div
              className={`max-w-[75%] rounded-2xl px-4 py-3 text-sm whitespace-pre-wrap leading-relaxed ${msg.role === "user"
                ? "bg-blue-600 text-white rounded-br-sm"
                : "bg-gray-800 text-gray-100 rounded-bl-sm"
                }`}
            >
              {/* Show a pulsing cursor while the assistant message is empty and streaming */}
              {msg.role === "assistant" && isStreaming && msg.content === "" ? (
                <span className="inline-block w-2 h-4 bg-gray-400 animate-pulse align-middle" />
              ) : (
                msg.content
              )}
            </div>
          </div>
        ))}
        <div ref={bottomRef} />
      </main>

      {/* Input bar */}
      <form
        onSubmit={handleSubmit}
        className="border-t border-gray-800 px-4 py-4 flex gap-3 shrink-0"
      >
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          disabled={isStreaming}
          placeholder={isStreaming ? "Thinking…" : "Type a message…"}
          className="flex-1 bg-gray-800 rounded-xl px-4 py-3 text-sm placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
        />
        <button
          type="submit"
          disabled={!input.trim() || isStreaming}
          className="bg-blue-600 hover:bg-blue-500 disabled:opacity-40 disabled:cursor-not-allowed text-white rounded-xl px-5 py-3 text-sm font-medium transition-colors"
        >
          Send
        </button>
      </form>
    </div>
  );
}
