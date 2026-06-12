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
// The HF Datasets Server returns rows inside a `rows` array where each element
// has a `row` object — we flatten that here for convenience.
// SQuAD schema: https://huggingface.co/datasets/rajpurkar/squad
type Row struct {
	Title    string `json:"title"`
	Context  string `json:"context"`
	Question string `json:"question"`
}

// datasetsResponse is the raw JSON shape returned by the Datasets Server.
// We only decode the fields we need; extra fields are silently ignored by json.Unmarshal.
type datasetsResponse struct {
	// Rows is a slice of structs — []T in Go, where T is an anonymous struct.
	// Anonymous structs are handy for one-off JSON shapes you won't reuse.
	Rows []struct {
		Row Row `json:"row"`
	} `json:"rows"`
}

// Load fetches up to limit rows of the given dataset+config from the HF Datasets Server.
// It returns them as a flat []Row ready for keyword retrieval.
func Load(ctx context.Context, dataset, config string, limit int) ([]Row, error) {
	// url.Values is a map[string][]string — use it to build query strings safely.
	// url.Values.Encode() percent-encodes each value, preventing injection.
	// https://pkg.go.dev/net/url#Values
	params := url.Values{}
	params.Set("dataset", dataset)
	params.Set("config", config)
	params.Set("split", "train")
	params.Set("offset", "0")
	params.Set("limit", fmt.Sprintf("%d", limit))

	endpoint := datasetsBaseURL + "?" + params.Encode()

	client := &http.Client{Timeout: 30 * time.Second}

	// http.NewRequestWithContext binds the context so this request is cancelled
	// if ctx is cancelled (e.g. the server shuts down before the load finishes).
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating dataset request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching dataset: %w", err)
	}
	// defer schedules resp.Body.Close() to run when Load returns.
	// This is canonical Go resource cleanup — no need for try/finally.
	// https://go.dev/tour/flowcontrol/12
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("datasets API %d: %s", resp.StatusCode, string(body))
	}

	var result datasetsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding dataset response: %w", err)
	}

	// Pre-allocate the slice with make([]T, 0, capacity).
	// The second arg (0) is the initial length; the third (len) is the capacity.
	// Pre-allocating avoids repeated reallocation as we append.
	// https://go.dev/tour/moretypes/13
	rows := make([]Row, 0, len(result.Rows))
	for _, r := range result.Rows {
		rows = append(rows, r.Row)
	}
	return rows, nil
}
