package compat

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	searchpkg "github.com/alessandrojcm/pampax-go/internal/search"
)

type fixtureSearchOutput struct {
	Query   string `json:"query"`
	Results []struct {
		Path string `json:"path"`
		Meta struct {
			ID    string  `json:"id"`
			Score float64 `json:"score"`
		} `json:"meta"`
	} `json:"results"`
}

type baselineMetrics struct {
	GoCompatibilityRequirements struct {
		MinimumAcceptableMetrics struct {
			BasePrecisionAt1 float64 `json:"base_precision_at_1"`
			BaseMRRAt5       float64 `json:"base_mrr_at_5"`
			BaseNDCGAt10     float64 `json:"base_ndcg_at_10"`
		} `json:"minimum_acceptable_metrics"`
	} `json:"go_compatibility_requirements"`
}

type unitQueryProvider struct{}

func (unitQueryProvider) GenerateEmbedding(_ string) ([]float64, error) {
	return []float64{1, 0}, nil
}

func (unitQueryProvider) GetDimensions() int { return 2 }
func (unitQueryProvider) GetName() string    { return "compat-test" }

func TestNodeFixtureSearchOrderingAndScores(t *testing.T) {
	searchDir := filepath.Join("..", "fixtures", "small", "search_outputs")
	queryFiles, err := filepath.Glob(filepath.Join(searchDir, "query_*.json"))
	if err != nil {
		t.Fatalf("glob fixture search outputs: %v", err)
	}
	if len(queryFiles) == 0 {
		t.Fatalf("no query_*.json files found in %s", searchDir)
	}

	provider := unitQueryProvider{}
	for _, queryPath := range queryFiles {
		queryPath := queryPath
		t.Run(filepath.Base(queryPath), func(t *testing.T) {
			raw, err := os.ReadFile(queryPath)
			if err != nil {
				t.Fatalf("read fixture search output: %v", err)
			}

			var fixture fixtureSearchOutput
			if err := json.Unmarshal(raw, &fixture); err != nil {
				t.Fatalf("unmarshal fixture search output: %v", err)
			}
			if len(fixture.Results) == 0 {
				t.Fatal("fixture search output has no results")
			}

			candidates := make([]searchpkg.Candidate, 0, len(fixture.Results))
			expectedScores := make(map[string]float64, len(fixture.Results))
			expectedOrder := make([]string, 0, len(fixture.Results))
			for _, item := range fixture.Results {
				score := item.Meta.Score
				if score > 1 {
					score = 1
				}
				if score < -1 {
					score = -1
				}

				candidates = append(candidates, searchpkg.Candidate{
					ID:   item.Meta.ID,
					Path: item.Path,
					Embedding: []float64{
						score,
						math.Sqrt(math.Max(0, 1-score*score)),
					},
				})
				expectedOrder = append(expectedOrder, item.Meta.ID)
				expectedScores[item.Meta.ID] = score
			}

			results, err := searchpkg.Search(fixture.Query, searchpkg.SearchOptions{
				Provider:   provider,
				Candidates: candidates,
				Limit:      10,
			})
			if err != nil {
				t.Fatalf("search stub execution failed: %v", err)
			}

			max := minInt(10, len(expectedOrder))
			if len(results) != max {
				t.Fatalf("result count mismatch: got %d, want %d", len(results), max)
			}

			for i := 0; i < max; i++ {
				if results[i].ID != expectedOrder[i] {
					t.Fatalf("result ordering mismatch at %d: got %q, want %q", i, results[i].ID, expectedOrder[i])
				}
				if math.Abs(results[i].Score-expectedScores[results[i].ID]) > 0.01 {
					t.Fatalf("score mismatch for %q: got %.6f, want %.6f", results[i].ID, results[i].Score, expectedScores[results[i].ID])
				}
			}
		})
	}
}

func TestNodeBaselineMetricsSanity(t *testing.T) {
	baselinePath := filepath.Join("..", "baselines", "node_baseline_2026-01-28.json")
	raw, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatalf("read baseline file: %v", err)
	}

	var baseline baselineMetrics
	if err := json.Unmarshal(raw, &baseline); err != nil {
		t.Fatalf("unmarshal baseline file: %v", err)
	}

	minMetrics := baseline.GoCompatibilityRequirements.MinimumAcceptableMetrics
	checks := []struct {
		name  string
		value float64
		min   float64
	}{
		{name: "base_precision_at_1", value: minMetrics.BasePrecisionAt1, min: 0.75},
		{name: "base_mrr_at_5", value: minMetrics.BaseMRRAt5, min: 0.75},
		{name: "base_ndcg_at_10", value: minMetrics.BaseNDCGAt10, min: 0.82},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if check.value < check.min {
				t.Fatalf("baseline metric %s too low: got %.3f, need >= %.3f", check.name, check.value, check.min)
			}
		})
	}

	if minMetrics.BasePrecisionAt1 != 0.75 || minMetrics.BaseMRRAt5 != 0.75 {
		t.Fatalf("unexpected baseline reference values: precision=%.3f mrr=%.3f", minMetrics.BasePrecisionAt1, minMetrics.BaseMRRAt5)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestNodeFixtureSearchFilesAreWellFormed(t *testing.T) {
	searchDir := filepath.Join("..", "fixtures", "small", "search_outputs")
	for i := 1; i <= 4; i++ {
		path := filepath.Join(searchDir, fmt.Sprintf("query_%03d.json", i))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected fixture file %s: %v", path, err)
		}
	}
}
