package main

import (
	"fmt"

	"github.com/alessandrojcm/pampax-go/internal/config"
	"github.com/alessandrojcm/pampax-go/internal/providers"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type reindexOptions struct {
	provider      string
	encryptionKey string
	encrypt       string
	project       string
	directory     string
}

func newCommandLogger(_ *cobra.Command) *zerolog.Logger {
	return &log.Logger
}

func addReindexFlags(cmd *cobra.Command, opts *reindexOptions) {
	cmd.Flags().StringVarP(&opts.provider, "provider", "p", "auto", "embedding provider")
	cmd.Flags().StringVar(&opts.encryptionKey, "encryption-key", "", "base64 or hex encryption key")
	cmd.Flags().StringVar(&opts.encrypt, "encrypt", "on", "encrypt chunk payloads (on|off)")
	cmd.Flags().StringVar(&opts.project, "project", "", "alias for project path")
	cmd.Flags().StringVar(&opts.directory, "directory", "", "alias for project directory")
}

func buildProviderConfig(cfg *config.Config) providers.FactoryConfig {
	if cfg == nil {
		return providers.FactoryConfig{}
	}

	return providers.FactoryConfig{
		OpenAIAPIKey:         cfg.OpenAIAPIKey,
		OpenAIBaseURL:        cfg.OpenAIBaseURL,
		OpenAIEmbeddingModel: cfg.OpenAIEmbeddingModel,
		TransformersModel:    cfg.TransformersModel,
		OllamaBaseURL:        cfg.OllamaBaseURL,
		OllamaModel:          cfg.OllamaModel,
		CohereAPIKey:         cfg.CohereAPIKey,
		CohereModel:          cfg.CohereModel,
		Dimensions:           cfg.Dimensions,
	}
}

func resolveProvider(requested string, cfg *config.Config) (providers.EmbeddingProvider, error) {
	provider, err := providers.NewEmbeddingProvider(requested, buildProviderConfig(cfg))
	if err != nil {
		return nil, fmt.Errorf("resolve embedding provider: %w", err)
	}

	return provider, nil
}

func runReindexScaffold(cmd *cobra.Command, args []string, commandName string, opts *reindexOptions, globals *globalOptions) error {
	if err := validateToggle("encrypt", opts.encrypt); err != nil {
		return err
	}

	provider, err := resolveProvider(opts.provider, globals.config)
	if err != nil {
		return err
	}

	targetPath := resolvePath(args, opts.project, opts.directory)
	newCommandLogger(cmd).Info().
		Str("command", commandName).
		Str("path", targetPath).
		Str("provider", opts.provider).
		Str("provider_name", provider.GetName()).
		Int("provider_dimensions", provider.GetDimensions()).
		Str("encrypt", opts.encrypt).
		Msg(commandName + " scaffold")

	return nil
}
