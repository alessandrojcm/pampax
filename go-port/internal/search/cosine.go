package search

import "math"

// CosineSimilarity computes cosine similarity for two vectors.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}

	var dotProduct float64
	var normA float64
	var normB float64

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	score := dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
	if score > 1 {
		return 1
	}
	if score < -1 {
		return -1
	}

	return score
}
