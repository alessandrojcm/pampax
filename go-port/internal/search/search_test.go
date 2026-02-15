package search

import (
	"fmt"
	"testing"
)

type providerStub struct {
	embedding []float64
	err       error
}

func (p providerStub) GenerateEmbedding(_ string) ([]float64, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.embedding, nil
}

func (p providerStub) GetDimensions() int {
	return len(p.embedding)
}

func (p providerStub) GetName() string {
	return "stub"
}

func TestSearchRanksByCosineAndLimits(t *testing.T) {
	provider := providerStub{embedding: []float64{1, 0}}

	results, err := Search("auth", SearchOptions{
		Provider: provider,
		Candidates: []Candidate{
			{ID: "a", Path: "a.go", Embedding: []float64{0.9, 0.1}},
			{ID: "b", Path: "b.go", Embedding: []float64{0.6, 0.8}},
			{ID: "c", Path: "c.go", Embedding: []float64{-1, 0}},
			{ID: "skip", Path: "skip.go", Embedding: []float64{1, 0, 0}},
		},
		Limit: 2,
	})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].ID != "a" {
		t.Fatalf("first result ID = %q, want %q", results[0].ID, "a")
	}
	if results[1].ID != "b" {
		t.Fatalf("second result ID = %q, want %q", results[1].ID, "b")
	}
}

func TestSearchValidation(t *testing.T) {
	if _, err := Search("  ", SearchOptions{Provider: providerStub{embedding: []float64{1}}}); err == nil {
		t.Fatal("expected error for empty query")
	}

	if _, err := Search("query", SearchOptions{}); err == nil {
		t.Fatal("expected error for missing provider")
	}

	if _, err := Search("query", SearchOptions{Provider: providerStub{err: fmt.Errorf("boom")}}); err == nil {
		t.Fatal("expected provider embedding error")
	}
}

func TestSearchDefaultLimit(t *testing.T) {
	provider := providerStub{embedding: []float64{1, 0}}

	candidates := make([]Candidate, 0, 12)
	for i := 0; i < 12; i++ {
		candidates = append(candidates, Candidate{
			ID:        fmt.Sprintf("id-%02d", i),
			Path:      fmt.Sprintf("f-%02d.go", i),
			Embedding: []float64{float64(12 - i), 0},
		})
	}

	results, err := Search("query", SearchOptions{Provider: provider, Candidates: candidates})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if len(results) != 10 {
		t.Fatalf("expected default 10 results, got %d", len(results))
	}
}
