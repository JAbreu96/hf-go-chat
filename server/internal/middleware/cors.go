// Package middleware provides HTTP handler wrappers (middleware).
// In Go, middleware is a function that takes an http.Handler and returns an http.Handler,
// letting you compose any handler with additional behaviour.
package middleware

import "net/http"

// Packages available for your implementation.
var (
	_ = http.HandlerFunc(nil) // a function type that implements http.Handler
	_ = http.MethodOptions    // the string "OPTIONS"
)

// CORS wraps next with permissive CORS headers so the Next.js dev server on :3000
// can call the Go server on :8080 without browser preflight errors.
//
// EXERCISE — implement this function.
//
// Concepts practiced:
//   - http.Handler interface: any type with ServeHTTP(ResponseWriter, *Request) satisfies it.
//     No `implements` keyword — Go interfaces are implicit.
//     https://pkg.go.dev/net/http#Handler
//   - http.HandlerFunc: a named function type defined as:
//       type HandlerFunc func(ResponseWriter, *Request)
//     It has a ServeHTTP method, so it satisfies http.Handler.
//     This is Go's "function that implements an interface" pattern.
//     https://pkg.go.dev/net/http#HandlerFunc
//   - Middleware signature: func(next http.Handler) http.Handler — the outer function
//     receives the handler to wrap; the returned handler adds behaviour around it.
//   - w.Header().Set(key, value): sets a response header before writing the body.
//   - next.ServeHTTP(w, r): delegates to the wrapped handler.
//
// Steps:
//  1. Return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { ... })
//     Inside the function:
//  2. Call w.Header().Set three times:
//       "Access-Control-Allow-Origin"  → "*"
//       "Access-Control-Allow-Methods" → "POST, GET, OPTIONS"
//       "Access-Control-Allow-Headers" → "Content-Type"
//  3. If r.Method == http.MethodOptions (the browser preflight), write status 204
//     with w.WriteHeader(http.StatusNoContent) and return immediately.
//  4. Otherwise call next.ServeHTTP(w, r) to hand off to the wrapped handler.
func CORS(next http.Handler) http.Handler {
	panic("not implemented")
}
