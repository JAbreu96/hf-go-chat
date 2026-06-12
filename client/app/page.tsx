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
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [isStreaming, setIsStreaming] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to the latest message after every render where messages changed.
  // useEffect(fn, [dep]) runs fn after the DOM updates whenever dep changes.
  // https://react.dev/reference/react/useEffect
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const text = input.trim();
    if (!text || isStreaming) return;

    const userMsg: Message = { id: crypto.randomUUID(), role: "user", content: text };
    const assistantMsg: Message = { id: crypto.randomUUID(), role: "assistant", content: "" };

    setMessages((prev) => [...prev, userMsg, assistantMsg]);
    setInput("");
    setIsStreaming(true);

    try {
      // POST to the Next.js Route Handler — same origin, no CORS issue.
      // The Route Handler proxies to the Go server server-side.
      const res = await fetch("/api/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        // Send the full conversation history so the model has context.
        body: JSON.stringify({
          messages: [...messages, userMsg].map(({ role, content }) => ({ role, content })),
        }),
      });

      if (!res.ok || !res.body) {
        throw new Error(`HTTP ${res.status}`);
      }

      // res.body is a ReadableStream<Uint8Array>. getReader() locks the stream
      // and returns a ReadableStreamDefaultReader to pull chunks from it.
      // https://developer.mozilla.org/en-US/docs/Web/API/ReadableStream/getReader
      const reader = res.body.getReader();
      // TextDecoder converts Uint8Array bytes to a UTF-8 string.
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { value, done } = await reader.read();
        if (done) break;

        // { stream: true } tells the decoder not to flush its internal state
        // between calls — needed when a multi-byte character spans two chunks.
        buffer += decoder.decode(value, { stream: true });

        // Split on newlines. A chunk boundary may land mid-line, so we keep
        // the last (possibly incomplete) segment in the buffer.
        const lines = buffer.split("\n");
        buffer = lines.pop() ?? "";

        for (const line of lines) {
          const delta = parseDelta(line.trim());
          if (delta) {
            // Functional setState: pass a function so we always operate on the
            // latest state — avoids stale closure capturing old messages array.
            setMessages((prev) =>
              prev.map((m) =>
                m.id === assistantMsg.id ? { ...m, content: m.content + delta } : m
              )
            );
          }
        }
      }
    } catch (err) {
      setMessages((prev) =>
        prev.map((m) =>
          m.id === assistantMsg.id ? { ...m, content: `Error: ${String(err)}` } : m
        )
      );
    } finally {
      setIsStreaming(false);
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
              className={`max-w-[75%] rounded-2xl px-4 py-3 text-sm whitespace-pre-wrap leading-relaxed ${
                msg.role === "user"
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
