package main

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type reindexOptions struct {
	provider      string
	encryptionKey string
	encrypt       string
	project       string
	directory     string
}

func newCommandLogger(cmd *cobra.Command) *zerolog.Logger {
	logger := zerolog.New(cmd.OutOrStdout()).With().Timestamp().Logger()
	return &logger
}

func addReindexFlags(cmd *cobra.Command, opts *reindexOptions) {
	cmd.Flags().StringVarP(&opts.provider, "provider", "p", "auto", "embedding provider")
	cmd.Flags().StringVar(&opts.encryptionKey, "encryption-key", "", "base64 or hex encryption key")
	cmd.Flags().StringVar(&opts.encrypt, "encrypt", "on", "encrypt chunk payloads (on|off)")
	cmd.Flags().StringVar(&opts.project, "project", "", "alias for project path")
	cmd.Flags().StringVar(&opts.directory, "directory", "", "alias for project directory")
}

func runReindexScaffold(cmd *cobra.Command, args []string, commandName string, opts *reindexOptions) error {
	if err := validateToggle("encrypt", opts.encrypt); err != nil {
		return err
	}

	targetPath := resolvePath(args, opts.project, opts.directory)
	newCommandLogger(cmd).Info().
		Str("command", commandName).
		Str("path", targetPath).
		Str("provider", opts.provider).
		Str("encrypt", opts.encrypt).
		Msg(commandName + " scaffold")

	return nil
}
