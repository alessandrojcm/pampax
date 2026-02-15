package main

import "github.com/spf13/cobra"

func newInfoCommand(globals *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show index health and basic statistics",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			provider, err := resolveProvider("auto", globals.config)
			if err != nil {
				return err
			}

			newCommandLogger(cmd).Info().
				Str("command", "info").
				Str("provider", provider.GetName()).
				Int("dimensions", provider.GetDimensions()).
				Int("chunks", 0).
				Int("db_size", 0).
				Msg("info scaffold")
			return nil
		},
	}

	return cmd
}
