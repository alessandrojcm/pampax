package main

import "github.com/spf13/cobra"

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

func newSearchCommand(_ *globalOptions) *cobra.Command {
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

			query := args[0]
			targetPath := resolvePath(args[1:], opts.project, opts.directory)
			newCommandLogger(cmd).Info().
				Str("command", "search").
				Str("query", query).
				Str("path", targetPath).
				Int("limit", opts.limit).
				Str("provider", opts.provider).
				Msg("search scaffold")
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
