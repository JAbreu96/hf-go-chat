package rag

import (
	"sort"
	"strings"
)

// Retrieve returns the top-n most relevant contexts for query.
// Relevance is scored by counting how many query words appear in the row's context.
// This is intentionally simple — production RAG uses vector embeddings.
func Retrieve(rows []Row, query string, topN int) []string {
	// strings.Fields splits on any whitespace and returns a []string.
	// https://pkg.go.dev/strings#Fields
	tokens := strings.Fields(strings.ToLower(query))

	type scored struct {
		context string
		score   int
	}

	// Make a slice of scored structs — scored{} is a struct literal.
	// We pre-allocate with len(rows) since we score every row.
	results := make([]scored, 0, len(rows))

	for _, row := range rows {
		lower := strings.ToLower(row.Context)
		score := 0
		for _, t := range tokens {
			// strings.Contains reports whether substr is within s.
			// https://pkg.go.dev/strings#Contains
			if strings.Contains(lower, t) {
				score++
			}
		}
		if score > 0 {
			results = append(results, scored{context: row.Context, score: score})
		}
	}

	// sort.Slice sorts a slice in place using a less function.
	// The less function receives indices i and j and returns true if i should come before j.
	// We sort descending by score, so higher scores come first.
	// https://pkg.go.dev/sort#Slice
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// Cap topN at the number of results we actually have.
	if topN > len(results) {
		topN = len(results)
	}

	// Range over a slice: the blank identifier _ discards the index we don't need.
	// https://go.dev/ref/spec#Blank_identifier
	contexts := make([]string, 0, topN)
	for _, r := range results[:topN] {
		contexts = append(contexts, r.context)
	}
	return contexts
}
