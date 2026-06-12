// Package rag loads a Hugging Face dataset at startup and exposes it for retrieval.
// "RAG" = Retrieval-Augmented Generation: inject relevant context into the prompt
// so the model can answer questions grounded in your data.
package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// datasetsBaseURL is the HF Datasets Server API.
// No auth required for public datasets.
// https://huggingface.co/docs/datasets-server/en/quick_start
const datasetsBaseURL = "https://datasets-server.huggingface.co/rows"

// Row is one record from the SQuAD dataset.
type Row struct {
	Title    string `json:"title"`
	Context  string `json:"context"`
	Question string `json:"question"`
}

// datasetsResponse is the raw JSON shape returned by the Datasets Server.
type datasetsResponse struct {
	Rows []struct {
		Row Row `json:"row"`
	} `json:"rows"`
}

// Packages available for your implementation.
var (
	_ = url.Values{}               // map[string][]string for building query strings — https://pkg.go.dev/net/url#Values
	_ = http.NewRequestWithContext // creates an HTTP request bound to a context — https://pkg.go.dev/net/http#NewRequestWithContext
	_ = json.NewDecoder            // wraps an io.Reader for streaming JSON decoding — https://pkg.go.dev/encoding/json#NewDecoder
	_ = io.ReadAll                 // reads all bytes from an io.Reader — https://pkg.go.dev/io#ReadAll
	_ = fmt.Sprintf                // formats a string (useful for int → string conversion) — https://pkg.go.dev/fmt#Sprintf
	_ = time.Second                // time.Duration constant — https://pkg.go.dev/time#Second
)

// Load fetches up to limit rows of the given dataset+config from the HF Datasets Server.
// It returns them as a flat []Row ready for keyword retrieval.
//
// EXERCISE — implement this function.
//
// Concepts practiced:
//   - url.Values: a map[string][]string for building URL query strings without
//     manual string concatenation or injection risk.
//     https://pkg.go.dev/net/url#Values
//   - http.NewRequestWithContext: creates an outbound HTTP request bound to a
//     context so it respects cancellation.
//     https://pkg.go.dev/net/http#NewRequestWithContext
//   - http.Client.Do: executes the request and returns (*http.Response, error).
//     https://pkg.go.dev/net/http#Client.Do
//   - defer resp.Body.Close(): schedules body cleanup to run when Load returns.
//     Go's HTTP client requires the body to be closed to reuse the connection.
//     https://go.dev/tour/flowcontrol/12
//   - json.NewDecoder(r).Decode(&v): streams JSON from an io.Reader into a struct.
//     https://pkg.go.dev/encoding/json#Decoder.Decode
//   - make([]T, 0, n): pre-allocates a slice with capacity n.
//     https://go.dev/tour/moretypes/13
//   - fmt.Errorf("context: %w", err): wraps an error with additional context.
//     https://go.dev/blog/go1.13-errors
//
// Steps:
//  1. Build the query string using url.Values. Set these keys:
//       "dataset" → dataset, "config" → config, "split" → "train",
//       "offset" → "0", "limit" → fmt.Sprintf("%d", limit)
//     Produce the full endpoint: datasetsBaseURL + "?" + params.Encode()
//
//  2. Create an http.Client with a 30-second timeout.
//
//  3. Build a GET request with http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil).
//     Return a wrapped error if this fails.
//
//  4. Execute the request with client.Do(req). Return a wrapped error on failure.
//     Immediately defer resp.Body.Close().
//
//  5. If resp.StatusCode != http.StatusOK, read the body with io.ReadAll and
//     return fmt.Errorf("datasets API %d: %s", resp.StatusCode, string(body)).
//
//  6. Decode the response body into a datasetsResponse struct.
//     Return a wrapped error if decoding fails.
//
//  7. Pre-allocate rows := make([]Row, 0, len(result.Rows)).
//     Loop over result.Rows, appending each r.Row to rows.
//
//  8. Return rows, nil.
func Load(ctx context.Context, dataset, config string, limit int) ([]Row, error) {
	panic("not implemented")
}
