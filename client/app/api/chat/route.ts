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

const GO_BACKEND = process.env.GO_BACKEND_URL ?? "http://localhost:8080";

export async function POST(req: NextRequest) {
  // req.json() parses the request body. NextRequest extends the Web API Request,
  // adding Next.js helpers like .nextUrl, .cookies, and .geo.
  // https://nextjs.org/docs/app/api-reference/functions/next-request
  const body = await req.json();

  const upstream = await fetch(`${GO_BACKEND}/api/chat`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    // Next.js extends fetch with caching options. `cache: "no-store"` disables
    // the Next.js data cache so every request hits the Go server fresh.
    // https://nextjs.org/docs/app/api-reference/functions/fetch
    cache: "no-store",
  });

  if (!upstream.ok) {
    return new Response("upstream error", { status: upstream.status });
  }

  // Re-stream the Go server's SSE response back to the browser.
  // upstream.body is a ReadableStream<Uint8Array> — we pipe it directly without
  // buffering the whole response in memory.
  // https://developer.mozilla.org/en-US/docs/Web/API/ReadableStream
  return new Response(upstream.body, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      Connection: "keep-alive",
    },
  });
}
