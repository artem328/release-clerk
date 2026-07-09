package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/artem328/release-clerk/internal/config"
	"github.com/artem328/release-clerk/internal/pkg/git"
	"github.com/artem328/release-clerk/internal/version"
	"github.com/artem328/release-clerk/pkg/semver"
	"github.com/spf13/cobra"
)

var tagFlags = struct {
	Format string
}{}

type tagOut struct {
	Valid   bool
	Version version.Version
}

var tagCmd = &cobra.Command{
	Use:   "tag <tag>",
	Short: "Display version information for tag",
	Args:  argsWrapper(cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := loadConfig()
		if err != nil {
			return err
		}

		out, err := resolveVersionFromTag(cmd.Context(), conf, args[0])
		if err != nil {
			return err
		}

		switch tagFlags.Format {
		case "json":
			j, err := json.Marshal(out)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(os.Stdout, string(j))
		case "text":
			if out.Valid {
				_, _ = fmt.Fprintln(os.Stdout, "Tag:    ", out.Version.Tag)
				_, _ = fmt.Fprintln(os.Stdout, "Version:", out.Version.SemVer.String())
				_, _ = fmt.Fprintln(os.Stdout, "Commit: ", out.Version.Commit.FullHash)
			}
		default:
			return fmt.Errorf("unknown format: %s", tagFlags.Format)
		}

		return nil
	},
}

func resolveVersionFromTag(ctx context.Context, conf config.Config, tag string) (tagOut, error) {
	r, err := git.LocateRepo(ctx)
	if err != nil {
		return tagOut{}, err
	}

	raw, ok := strings.CutPrefix(tag, conf.TagPrefix)
	if !ok {
		return tagOut{Valid: false}, nil
	}

	ver, err := semver.Parse(raw)
	if err != nil {
		return tagOut{Valid: false}, nil
	}

	v, err := version.Resolve(r, ver, conf.TagPrefix)
	if err != nil {
		return tagOut{Valid: false}, nil
	}

	return tagOut{Valid: true, Version: v}, nil
}

func init() {
	tagCmd.Flags().StringVarP(&tagFlags.Format, "format", "f", "text", "Format to output (json|text)")

	rootCmd.AddCommand(tagCmd)
}
