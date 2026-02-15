package search

import (
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
	}{
		{
			name:     "identical vectors",
			a:        []float64{1, 2, 3},
			b:        []float64{1, 2, 3},
			expected: 1,
		},
		{
			name:     "orthogonal vectors",
			a:        []float64{1, 0},
			b:        []float64{0, 1},
			expected: 0,
		},
		{
			name:     "opposite vectors",
			a:        []float64{1, 0},
			b:        []float64{-1, 0},
			expected: -1,
		},
		{
			name:     "mismatched dimensions",
			a:        []float64{1, 0},
			b:        []float64{1, 0, 0},
			expected: 0,
		},
		{
			name:     "zero norm vector",
			a:        []float64{0, 0},
			b:        []float64{1, 0},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CosineSimilarity(tt.a, tt.b)
			if math.Abs(got-tt.expected) > 1e-9 {
				t.Fatalf("CosineSimilarity() = %f, want %f", got, tt.expected)
			}
		})
	}
}
