package main

import "github.com/spf13/cobra"

func newInfoCommand(_ *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show index health and basic statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			newCommandLogger(cmd).Info().
				Str("command", "info").
				Str("provider", "stub").
				Int("chunks", 0).
				Int("db_size", 0).
				Msg("info scaffold")
			return nil
		},
	}

	return cmd
}
