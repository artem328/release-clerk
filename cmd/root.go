package cmd

import (
	"context"

	"github.com/artem328/release-clerk/internal/config"
	"github.com/spf13/cobra"
)

var globalFlags = struct {
	Config string
}{}

var rootCmd = &cobra.Command{
	Use:   "release-clerk",
	Short: "Release CLI",
}

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&globalFlags.Config,
		"config",
		"c",
		"",
		"release-clerk config location. By default trying to find one of .release-clerk.yaml, .release-clerk.yml in working directory",
	)
}

func loadConfig() (config.Config, error) {
	if globalFlags.Config != "" {
		return config.Load(globalFlags.Config)
	}

	return config.Discover()
}
