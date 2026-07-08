package cmd

import (
	"context"

	"github.com/artem328/release-clerk/internal/config"
	"github.com/spf13/cobra"
)

var globalFlags = struct {
	Config string
	Debug  bool
}{}

var rootCmd = &cobra.Command{
	Use:           "release-clerk",
	Short:         "Release CLI",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute(ctx context.Context) (*cobra.Command, error) {
	return rootCmd.ExecuteContextC(ctx)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&globalFlags.Config,
		"config",
		"c",
		"",
		"release-clerk config location. By default trying to find one of .release-clerk.yaml, .release-clerk.yml in working directory",
	)
	rootCmd.PersistentFlags().BoolVarP(&globalFlags.Debug, "debug", "d", false, "enable debug logging")
}

func loadConfig() (config.Config, error) {
	if globalFlags.Config != "" {
		return config.Load(globalFlags.Config)
	}

	return config.Discover()
}
