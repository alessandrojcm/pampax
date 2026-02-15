package main

import "github.com/spf13/cobra"

func newIndexCommand(_ *globalOptions) *cobra.Command {
	opts := &reindexOptions{}

	cmd := &cobra.Command{
		Use:   "index [path]",
		Short: "Index a project and produce .pampa artifacts",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReindexScaffold(cmd, args, "index", opts)
		},
	}

	addReindexFlags(cmd, opts)

	return cmd
}
