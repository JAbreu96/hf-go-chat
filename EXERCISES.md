# Exercise: next-01-route-handler

## What to implement

`client/app/api/chat/route.ts` — the POST Route Handler that proxies chat requests from the browser to the Go backend and re-streams the SSE response.

## Concepts covered

| Concept | Doc |
|---|---|
| Route Handler file convention (`route.ts`, named HTTP verb exports) | https://nextjs.org/docs/app/api-reference/file-conventions/route |
| `NextRequest` — Web API `Request` extended by Next.js | https://nextjs.org/docs/app/api-reference/functions/next-request |
| Server-only env vars (`process.env` in Route Handlers stays off the client bundle) | https://nextjs.org/docs/app/building-your-application/configuring/environment-variables |
| Next.js extended `fetch` with `cache: "no-store"` (opt out of data cache) | https://nextjs.org/docs/app/api-reference/functions/fetch |
| `ReadableStream` — streaming body piped through without buffering | https://developer.mozilla.org/en-US/docs/Web/API/ReadableStream |
| `Response` constructor — first arg is `BodyInit`, which accepts `ReadableStream` | https://developer.mozilla.org/en-US/docs/Web/API/Response/Response |
| SSE response headers (`text/event-stream`, `no-cache`, `keep-alive`) | https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events |

## Implementation checklist

- [ ] `await req.json()` to get the request body
- [ ] `fetch(${GO_BACKEND}/api/chat, { method, headers, body, cache: "no-store" })`
- [ ] Return error response if `!upstream.ok`
- [ ] Return `new Response(upstream.body, { headers: { SSE headers } })`

## How to test

1. Start the Go server: `cd server && go run ./cmd/main.go`
2. Start Next.js: `cd client && npm run dev`
3. Open http://localhost:3000 and send a message — tokens should stream in
4. Or hit the route directly:
   ```bash
   curl -N -X POST http://localhost:3000/api/chat \
     -H 'Content-Type: application/json' \
     -d '{"messages":[{"role":"user","content":"What is machine learning?"}]}'
   ```

## Solution

See `main` branch — `client/app/api/chat/route.ts`
