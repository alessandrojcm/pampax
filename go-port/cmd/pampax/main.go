package main

import (
	"fmt"
	"os"

	"github.com/alessandrojcm/pampax-go/internal/config"
	"github.com/alessandrojcm/pampax-go/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type globalOptions struct {
	pretty     bool
	verbose    bool
	configFile string
	config     *config.Config
}

func main() {
	utils.SetupLogger(utils.LoggingOptions{Writer: os.Stderr})

	rootCmd := NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("command failed")
		os.Exit(1)
	}
}

func NewRootCommand() *cobra.Command {
	opts := &globalOptions{}

	rootCmd := &cobra.Command{
		Use:           "pampax",
		Short:         "PAMPAX semantic code memory CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			loadedConfig, err := config.Load(opts.configFile)
			if err != nil {
				return err
			}
			opts.config = loadedConfig

			utils.SetupLogger(utils.LoggingOptions{
				Pretty:  opts.pretty,
				Verbose: opts.verbose,
				Writer:  cmd.OutOrStdout(),
			})

			return nil
		},
	}

	rootCmd.PersistentFlags().BoolVar(&opts.pretty, "pretty", false, "enable pretty console logging")
	rootCmd.PersistentFlags().StringVar(&opts.configFile, "config", "", "path to config file")
	rootCmd.PersistentFlags().BoolVar(&opts.verbose, "verbose", false, "enable verbose logs")

	rootCmd.AddCommand(newIndexCommand(opts))
	rootCmd.AddCommand(newUpdateCommand(opts))
	rootCmd.AddCommand(newSearchCommand(opts))
	rootCmd.AddCommand(newInfoCommand(opts))

	return rootCmd
}

func resolvePath(args []string, project string, directory string) string {
	if len(args) > 0 && args[0] != "" {
		return args[0]
	}
	if project != "" {
		return project
	}
	if directory != "" {
		return directory
	}
	return "."
}

func validateToggle(flag string, value string) error {
	if value == "on" || value == "off" {
		return nil
	}
	return fmt.Errorf("invalid %s value %q: must be one of [on, off]", flag, value)
}
