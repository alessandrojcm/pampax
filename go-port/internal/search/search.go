package search

import (
	"fmt"
	"sort"
	"strings"
)

// EmbeddingProvider defines the search embedding dependency.
type EmbeddingProvider interface {
	GenerateEmbedding(text string) ([]float64, error)
	GetDimensions() int
	GetName() string
}

// Candidate represents a searchable item with an existing embedding.
type Candidate struct {
	ID        string
	Path      string
	Embedding []float64
}

// SearchOptions captures Stage 4B search behavior inputs.
type SearchOptions struct {
	Provider    EmbeddingProvider
	Candidates  []Candidate
	Limit       int
	Hybrid      string
	BM25        string
	SymbolBoost string
	Reranker    string
}

// Result is a search hit sorted by descending score.
type Result struct {
	ID    string
	Path  string
	Score float64
}

// Search returns top-k results using cosine similarity.
// Stage 4B intentionally ignores BM25/hybrid/reranker options.
func Search(query string, options SearchOptions) ([]Result, error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}
	if options.Provider == nil {
		return nil, fmt.Errorf("embedding provider is required")
	}

	queryEmbedding, err := options.Provider.GenerateEmbedding(trimmedQuery)
	if err != nil {
		return nil, fmt.Errorf("generate query embedding: %w", err)
	}

	results := make([]Result, 0, len(options.Candidates))
	for _, candidate := range options.Candidates {
		if len(candidate.Embedding) == 0 || len(candidate.Embedding) != len(queryEmbedding) {
			continue
		}

		results = append(results, Result{
			ID:    candidate.ID,
			Path:  candidate.Path,
			Score: CosineSimilarity(queryEmbedding, candidate.Embedding),
		})
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			if results[i].Path == results[j].Path {
				return results[i].ID < results[j].ID
			}
			return results[i].Path < results[j].Path
		}
		return results[i].Score > results[j].Score
	})

	limit := options.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > len(results) {
		limit = len(results)
	}

	return results[:limit], nil
}
