// Package middleware provides HTTP handler wrappers (middleware).
// In Go, middleware is a function that takes an http.Handler and returns an http.Handler,
// letting you wrap any handler with additional behavior.
package middleware

import "net/http"

// CORS wraps next with permissive CORS headers so the Next.js dev server on :3000
// can call the Go server on :8080 without browser preflight errors.
// In production you would tighten the Origin allowlist.
//
// http.Handler is an interface with one method: ServeHTTP(ResponseWriter, *Request).
// Any type that implements that method satisfies http.Handler — no `implements` keyword.
// https://pkg.go.dev/net/http#Handler
func CORS(next http.Handler) http.Handler {
	// http.HandlerFunc is a named function type defined in net/http:
	//   type HandlerFunc func(ResponseWriter, *Request)
	// It has a ServeHTTP method, so it satisfies the http.Handler interface.
	// This is Go's "function type that implements an interface" pattern.
	// https://pkg.go.dev/net/http#HandlerFunc
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// OPTIONS is the browser preflight request — respond immediately and return.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// next.ServeHTTP calls the wrapped handler.
		next.ServeHTTP(w, r)
	})
}
