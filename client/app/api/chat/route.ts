// Route Handler: app/api/chat/route.ts → handles POST /api/chat
//
// In the App Router a file named route.ts inside app/ exports named functions
// matching HTTP verbs (GET, POST, PUT…). There is no default export, no Express-style
// (req, res) — just the Web standard Request / Response APIs.
// https://nextjs.org/docs/app/api-reference/file-conventions/route
//
// WHY this proxy exists:
//   1. The Go server URL lives in a server-only env var (GO_BACKEND_URL) — keeping it
//      out of the browser bundle prevents accidental exposure.
//   2. We avoid CORS issues by having same-origin requests from the browser hit Next.js,
//      which then calls Go server-side.

import { NextRequest } from "next/server";

// process.env is available here because Route Handlers run on the server (Node.js),
// never in the browser. Server-only env vars stay out of the client bundle.
// https://nextjs.org/docs/app/building-your-application/configuring/environment-variables
const GO_BACKEND = process.env.GO_BACKEND_URL ?? "http://localhost:8080";

// EXERCISE: implement this Route Handler.
//
// This POST handler receives the chat request from the browser, proxies it to the
// Go backend, and re-streams the Server-Sent Events (SSE) response back as-is.
//
// Step 1 — Parse the request body.
//   const body = await req.json()
//   NextRequest.json() returns the parsed JSON body as a plain object.
//   https://nextjs.org/docs/app/api-reference/functions/next-request
//
// Step 2 — Forward the request to the Go backend.
//   Call the global fetch() — Next.js extends it with caching options.
//   Use these options:
//     method: "POST"
//     headers: { "Content-Type": "application/json" }
//     body: JSON.stringify(body)
//     cache: "no-store"   ← opt out of Next.js data cache so every call hits Go fresh
//   https://nextjs.org/docs/app/api-reference/functions/fetch
//
// Step 3 — Handle upstream errors.
//   if (!upstream.ok) return new Response("upstream error", { status: upstream.status })
//   https://developer.mozilla.org/en-US/docs/Web/API/Response
//
// Step 4 — Re-stream the SSE body back to the browser.
//   upstream.body is a ReadableStream<Uint8Array>. Pass it directly to new Response()
//   as the first argument — no buffering needed, the bytes flow through.
//   https://developer.mozilla.org/en-US/docs/Web/API/ReadableStream
//
//   The response must include these headers so the browser treats it as SSE:
//     "Content-Type": "text/event-stream"
//     "Cache-Control": "no-cache"
//     "Connection": "keep-alive"
//
// How to test once implemented:
//   1. Start Go server: cd server && go run ./cmd/main.go
//   2. Start Next.js: cd client && npm run dev
//   3. Open http://localhost:3000 and send a message — tokens should stream in.
//   OR test the route directly:
//     curl -N -X POST http://localhost:3000/api/chat \
//       -H 'Content-Type: application/json' \
//       -d '{"messages":[{"role":"user","content":"Hello"}]}'
export async function POST(req: NextRequest) {
  const body = await req.json();

  const REQ_BODY = {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(body),
    cache: "no-store" as const
  }

  const res = await fetch(`${GO_BACKEND}/api/chat`, REQ_BODY);

  if (!res.ok) {
    return new Response("upstream error", { status: res.status })
  }

  const sse_response = new Response(res.body, {
    headers: {
      "Content-Type": "application/json",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive"
    }
  })

  return sse_response;
}
