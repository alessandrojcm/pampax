package main

import (
	"fmt"

	searchpkg "github.com/alessandrojcm/pampax-go/internal/search"
	"github.com/spf13/cobra"
)

type searchOptions struct {
	provider    string
	project     string
	directory   string
	limit       int
	pathGlobs   []string
	tags        []string
	languages   []string
	reranker    string
	hybrid      string
	bm25        string
	symbolBoost string
}

func newSearchCommand(globals *globalOptions) *cobra.Command {
	opts := &searchOptions{}

	cmd := &cobra.Command{
		Use:   "search <query> [path]",
		Short: "Search indexed chunks",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateToggle("hybrid", opts.hybrid); err != nil {
				return err
			}
			if err := validateToggle("bm25", opts.bm25); err != nil {
				return err
			}
			if err := validateToggle("symbol_boost", opts.symbolBoost); err != nil {
				return err
			}

			provider, err := resolveProvider(opts.provider, globals.config)
			if err != nil {
				return err
			}

			query := args[0]
			targetPath := resolvePath(args[1:], opts.project, opts.directory)

			candidates, err := buildSearchStubCandidates(provider)
			if err != nil {
				return err
			}

			results, err := searchpkg.Search(query, searchpkg.SearchOptions{
				Provider:    provider,
				Candidates:  candidates,
				Limit:       opts.limit,
				Hybrid:      opts.hybrid,
				BM25:        opts.bm25,
				SymbolBoost: opts.symbolBoost,
				Reranker:    opts.reranker,
			})
			if err != nil {
				return err
			}

			logger := newCommandLogger(cmd).Info().
				Str("command", "search").
				Str("query", query).
				Str("path", targetPath).
				Int("limit", opts.limit).
				Str("provider", opts.provider).
				Str("provider_name", provider.GetName()).
				Int("provider_dimensions", provider.GetDimensions()).
				Int("result_count", len(results))

			if len(results) > 0 {
				logger = logger.
					Str("top_result_id", results[0].ID).
					Float64("top_result_score", results[0].Score)
			}

			logger.Msg("search scaffold")
			return nil
		},
	}

	cmd.Flags().IntVarP(&opts.limit, "limit", "k", 10, "maximum number of results")
	cmd.Flags().IntVar(&opts.limit, "top", 10, "maximum number of results (alias for --limit)")
	cmd.Flags().StringVarP(&opts.provider, "provider", "p", "auto", "embedding provider")
	cmd.Flags().StringVar(&opts.project, "project", "", "alias for project path")
	cmd.Flags().StringVar(&opts.directory, "directory", "", "alias for project directory")
	cmd.Flags().StringArrayVar(&opts.pathGlobs, "path_glob", nil, "filter by file path glob")
	cmd.Flags().StringArrayVar(&opts.tags, "tags", nil, "filter by tags")
	cmd.Flags().StringArrayVar(&opts.languages, "lang", nil, "filter by language")
	cmd.Flags().StringVar(&opts.reranker, "reranker", "off", "reranker mode (off|transformers|api)")
	cmd.Flags().StringVar(&opts.hybrid, "hybrid", "on", "hybrid search mode (on|off)")
	cmd.Flags().StringVar(&opts.bm25, "bm25", "on", "BM25 candidate generation (on|off)")
	cmd.Flags().StringVar(&opts.symbolBoost, "symbol_boost", "on", "symbol-aware ranking boost (on|off)")

	return cmd
}

func buildSearchStubCandidates(provider searchpkg.EmbeddingProvider) ([]searchpkg.Candidate, error) {
	texts := []string{
		"repository overview and architecture",
		"authentication and login flow",
		"embedding provider configuration",
		"database schema and migrations",
		"chunk storage and encryption",
		"search ranking and relevance scoring",
	}

	candidates := make([]searchpkg.Candidate, 0, len(texts))
	for i, text := range texts {
		embedding, err := provider.GenerateEmbedding(text)
		if err != nil {
			return nil, fmt.Errorf("generate search stub candidate embedding: %w", err)
		}

		candidates = append(candidates, searchpkg.Candidate{
			ID:        fmt.Sprintf("stub-%02d", i+1),
			Path:      fmt.Sprintf("stub/doc-%02d.md", i+1),
			Embedding: embedding,
		})
	}

	return candidates, nil
}
