// Package handler contains the HTTP handler functions registered on the mux.
// Each handler is a func(http.ResponseWriter, *http.Request) — Go's standard
// handler signature. No framework, no magic — just the stdlib.
package handler

import "net/http"

// Health handles GET /health.
// http.ResponseWriter is an interface; w.WriteHeader sets the status code.
// https://pkg.go.dev/net/http#ResponseWriter
func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
