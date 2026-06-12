// Package tools contains the tool implementations the agent can invoke.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const braveSearchURL = "https://api.search.brave.com/res/v1/web/search"

// braveResponse is the subset of the Brave Search API response we need.
// Full schema: https://api.search.brave.com/app/documentation/web-search/responses
type braveResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			URL         string `json:"url"`
		} `json:"results"`
	} `json:"web"`
}

// Packages available for your implementation.
var (
	_ = url.Values{}
	_ = http.NewRequestWithContext
	_ = json.NewDecoder
	_ = io.ReadAll
	_ = fmt.Sprintf
	_ = time.Second
	_ = strings.Builder{}
)

// Search calls the Brave Search API and returns the top results as a
// human-readable string suitable for injecting into a chat message.
// API docs: https://api.search.brave.com/app/documentation/web-search/get-started
//
// EXERCISE — implement this function.
//
// Concepts practiced:
//   - url.Values for query string encoding (same as loader.go but now with a custom header).
//     https://pkg.go.dev/net/url#Values
//   - Setting HTTP request headers: req.Header.Set("Key", "value").
//     https://pkg.go.dev/net/http#Header.Set
//   - strings.Builder: the idiomatic way to build a string incrementally
//     without allocating a new string on every concatenation.
//     https://pkg.go.dev/strings#Builder
//   - fmt.Fprintf(&sb, format, args...): writes a formatted string to any io.Writer,
//     including a *strings.Builder (which implements io.Writer via its Write method).
//     https://pkg.go.dev/fmt#Fprintf
//
// Steps:
//  1. Build query params with url.Values: set "q" → query, "count" → fmt.Sprintf("%d", count).
//     Endpoint: braveSearchURL + "?" + params.Encode()
//
//  2. Create an http.Client{Timeout: 15 * time.Second}.
//
//  3. Build a GET request with http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil).
//     Add two headers:
//       "X-Subscription-Token" → apiKey
//       "Accept"               → "application/json"
//
//  4. Execute the request. Defer resp.Body.Close().
//     If StatusCode != http.StatusOK, read body and return a formatted error.
//
//  5. Decode into a braveResponse struct.
//
//  6. If len(result.Web.Results) == 0, return "No web results found.", nil.
//
//  7. Build the result string using strings.Builder:
//     For each result (index i, zero-based), write:
//       fmt.Fprintf(&sb, "[%d] %s\n%s\nURL: %s\n\n", i+1, title, description, url)
//
//  8. Return sb.String(), nil.
func Search(ctx context.Context, apiKey, query string, count int) (string, error) {
	panic("not implemented")
}
