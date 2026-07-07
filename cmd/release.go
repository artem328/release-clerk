package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/artem328/release-clerk/internal/pkg/log"
	"github.com/artem328/release-clerk/internal/release"
	"github.com/spf13/cobra"
)

var releaseCmdFlags = struct {
	DryRun bool
	Format string
}{}

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Create a release",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := loadConfig()
		if err != nil {
			return err
		}

		logger := log.NewGenericLogger(os.Stderr, "", true)

		rel, err := release.Run(cmd.Context(), conf, releaseCmdFlags.DryRun, logger)
		if err != nil {
			return err
		}

		switch releaseCmdFlags.Format {
		case "json":
			j, err := json.Marshal(rel)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(os.Stdout, string(j))
		case "text":
			if rel.Released {
				_, _ = fmt.Fprintln(os.Stdout, "Tag:    ", rel.Tag)
				_, _ = fmt.Fprintln(os.Stdout, "Version:", rel.Version.SemVer.String())
				_, _ = fmt.Fprintln(os.Stdout, "Commit: ", rel.Version.Commit.FullHash)
			}
		default:
			return fmt.Errorf("unknown format: %s", tagFlags.Format)
		}

		return nil
	},
}

func init() {
	releaseCmd.Flags().BoolVar(
		&releaseCmdFlags.DryRun,
		"dry-run",
		false,
		"Don't apply any changes, just show what would be done",
	)
	releaseCmd.Flags().StringVarP(&releaseCmdFlags.Format, "format", "f", "text", "Format to output (json|text)")

	rootCmd.AddCommand(releaseCmd)
}
