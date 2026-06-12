// Package tools contains the tool implementations the agent can invoke.
// Each tool is a plain Go function that calls an external API and returns a string
// the model can read as context.
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

// Search calls the Brave Search API and returns the top results as a
// human-readable string suitable for injecting into a chat message.
// https://api.search.brave.com/app/documentation/web-search/get-started
func Search(ctx context.Context, apiKey, query string, count int) (string, error) {
	// url.Values encodes query parameters safely — spaces become %20, etc.
	// https://pkg.go.dev/net/url#Values
	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", count))

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		braveSearchURL+"?"+params.Encode(),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("creating brave request: %w", err)
	}
	req.Header.Set("X-Subscription-Token", apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("brave search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("brave API %d: %s", resp.StatusCode, string(body))
	}

	var result braveResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding brave response: %w", err)
	}

	if len(result.Web.Results) == 0 {
		return "No web results found.", nil
	}

	// strings.Builder is the idiomatic way to build a string incrementally.
	// It avoids allocating a new string on each concatenation.
	// https://pkg.go.dev/strings#Builder
	var sb strings.Builder
	for i, r := range result.Web.Results {
		fmt.Fprintf(&sb, "[%d] %s\n%s\nURL: %s\n\n", i+1, r.Title, r.Description, r.URL)
	}
	return sb.String(), nil
}
