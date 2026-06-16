package rag

import (
	"sort"
	"strings"
)

// Packages available for your implementation.
var (
	_ = strings.Fields   // splits on whitespace → []string — https://pkg.go.dev/strings#Fields
	_ = strings.ToLower  // lowercases a string — https://pkg.go.dev/strings#ToLower
	_ = strings.Contains // reports whether substr is in s — https://pkg.go.dev/strings#Contains
	_ = sort.Slice       // sorts a slice in place with a less func — https://pkg.go.dev/sort#Slice
)

// Retrieve returns the top-n most relevant Row contexts for the given query.
// Relevance is scored by counting how many query words appear in each row's Context field.
//
// EXERCISE — implement this function.
//
// Concepts practiced:
//   - strings.Fields: splits a string on whitespace into a []string of tokens.
//     https://pkg.go.dev/strings#Fields
//   - strings.ToLower: lowercases a string for case-insensitive comparison.
//     https://pkg.go.dev/strings#ToLower
//   - strings.Contains: reports whether a substring appears in a string.
//     https://pkg.go.dev/strings#Contains
//   - Anonymous struct: a struct type defined inline, useful for one-off groupings.
//     https://go.dev/ref/spec#Struct_types
//   - make([]T, 0, capacity): allocates a slice with a pre-set capacity to avoid
//     repeated reallocation as you append. https://go.dev/tour/moretypes/13
//   - append: adds elements to a slice, growing it if needed.
//     https://go.dev/tour/moretypes/15
//   - sort.Slice: sorts in-place using a comparison function you provide.
//     https://pkg.go.dev/sort#Slice
//   - Blank identifier _: discards the loop index when you only need the value.
//     https://go.dev/ref/spec#Blank_identifier
//
// Steps:
//  1. Tokenise the query: split it into words with strings.Fields after lowercasing.
//  2. Define an anonymous struct type `scored` with fields `context string` and `score int`.
//     Allocate a `results` slice of that type with make([]scored, 0, len(rows)).
//  3. Loop over rows. For each row, count how many tokens from step 1 appear in
//     strings.ToLower(row.Context) using strings.Contains.
//     If the score is greater than 0, append a scored{context: ..., score: ...} to results.
//  4. Sort results descending by score using sort.Slice.
//     The less function: results[i].score > results[j].score  (higher score = earlier).
//  5. Cap topN: if topN > len(results), set topN = len(results).
//  6. Build the return value: allocate a []string with make([]string, 0, topN),
//     then append results[k].context for k in 0..topN-1.
//     Hint: range over results[:topN] and use _ to discard the index.
//  7. Return the slice of context strings.

type scored struct {
	context string
	score   int
}

func Retrieve(rows []Row, query string, topN int) []string {

	split_words := strings.Fields(strings.ToLower(query))

	results := make([]scored, 0, len(rows))

	for _, row := range rows {
		lower_row := strings.ToLower(row.Context)

		var score int = 0
		for _, word := range split_words {
			if strings.Contains(lower_row, word) {
				score++
			}
		}

		if score > 0 {
			results = append(results, scored{
				context: row.Context,
				score:   score,
			})
		}
	}

	sort.Slice(results, func(a, b int) bool {
		return results[a].score > results[b].score
	})

	if topN > len(results) {
		topN = len(results)
	}

	out := make([]string, 0, topN)

	for _, r := range results[:topN] {

		out = append(out, r.context)
	}

	return out
}
