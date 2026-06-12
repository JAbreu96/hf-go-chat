# Exercise: next-02-streaming-ui

## What to implement

`client/app/page.tsx` — the `handleSubmit` function inside the `ChatPage` component.

This function is the core of the streaming chat UI: it posts the conversation to `/api/chat`,
reads the SSE response body as a stream, decodes tokens chunk by chunk, and appends
each token to the assistant message in real time.

`parseDelta` (the SSE line parser) is already provided — focus on the streaming fetch loop.

## Concepts covered

| Concept | Doc |
|---|---|
| `'use client'` — opts a module into Client Component rendering (runs in browser) | https://nextjs.org/docs/app/building-your-application/rendering/client-components |
| `useState<T>` — typed state; setter with functional update form `(prev) => ...` | https://react.dev/reference/react/useState |
| `useRef` — mutable ref that doesn't trigger re-renders; DOM node reference | https://react.dev/reference/react/useRef |
| `useEffect` — side effects after DOM updates; auto-scroll via `scrollIntoView` | https://react.dev/reference/react/useEffect |
| `fetch` — POST with JSON body to same-origin Route Handler | https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API/Using_Fetch |
| `ReadableStream.getReader()` — lock the stream and pull chunks | https://developer.mozilla.org/en-US/docs/Web/API/ReadableStream/getReader |
| `TextDecoder` with `{ stream: true }` — preserve multi-byte char state across chunks | https://developer.mozilla.org/en-US/docs/Web/API/TextDecoder/decode |
| Manual SSE line parsing (split on `\n`, keep incomplete tail in buffer) | https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events |
| Functional `setState` to avoid stale closures when streaming dozens of chunks | https://react.dev/reference/react/useState#updating-state-based-on-the-previous-state |
| `crypto.randomUUID()` — generate stable message IDs for React's `key` prop | https://developer.mozilla.org/en-US/docs/Web/API/Crypto/randomUUID |

## Implementation checklist

- [ ] `e.preventDefault()` + guard against empty input / double-submit
- [ ] Optimistically push `userMsg` + empty `assistantMsg` into `messages` state
- [ ] `setInput("")` + `setIsStreaming(true)` before the fetch
- [ ] `fetch("/api/chat", { method: "POST", ... })` with full conversation history
- [ ] `res.body.getReader()` + `new TextDecoder()`
- [ ] `while(true)` loop reading chunks until `done`
- [ ] `decoder.decode(value, { stream: true })` appended to buffer
- [ ] Split buffer on `"\n"`, keep the last segment for the next chunk
- [ ] `parseDelta` each line; functional `setMessages` appending token to `assistantMsg`
- [ ] `catch` updating `assistantMsg` to the error string
- [ ] `finally` clearing `setIsStreaming(false)`

## How to test

1. Start Go server: `cd server && go run ./cmd/main.go`
2. Start Next.js: `cd client && npm run dev`
3. Open http://localhost:3000
4. Type a message and hit Send — tokens should appear word by word in the chat bubble

## Solution

See `main` branch — `client/app/page.tsx`
