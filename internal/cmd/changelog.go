package cmd

import (
	"fmt"
	"os"

	"github.com/artem328/release-clerk/internal/changelog"
	"github.com/artem328/release-clerk/pkg/semver"
	"github.com/spf13/cobra"
)

var changelogFlags = struct {
	Version string
}{}

var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generates changelog for version",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := loadConfig()
		if err != nil {
			return err
		}

		var v semver.Version
		if changelogFlags.Version != "" {
			v, err = semver.Parse(changelogFlags.Version)
			if err != nil {
				return err
			}
		}

		cl, err := changelog.Generate(cmd.Context(), conf, changelog.Config{
			Version: v,
		})
		if err != nil {
			return err
		}

		_, _ = fmt.Fprint(os.Stdout, cl)

		return nil
	},
}

func init() {
	changelogCmd.Flags().StringVarP(&changelogFlags.Version, "version", "v", "", "The version to generate changelog for. If empty the last version is used")

	rootCmd.AddCommand(changelogCmd)
}
