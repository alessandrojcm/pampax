package main

import "github.com/spf13/cobra"

func newUpdateCommand(_ *globalOptions) *cobra.Command {
	opts := &reindexOptions{}

	cmd := &cobra.Command{
		Use:   "update [path]",
		Short: "Reindex a project (full reindex for Stage 1)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReindexScaffold(cmd, args, "update", opts)
		},
	}

	addReindexFlags(cmd, opts)

	return cmd
}
