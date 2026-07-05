package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/artem328/release-clerk/internal/pkg/git"
	"github.com/artem328/release-clerk/internal/version"
	"github.com/artem328/release-clerk/pkg/semver"
	"github.com/spf13/cobra"
)

var tagFlags = struct {
	Format string
}{}

var tagCmd = &cobra.Command{
	Use:   "tag <tag>",
	Short: "Display version information for tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := loadConfig()
		if err != nil {
			return err
		}

		r, err := git.LocateRepo(cmd.Context())
		if err != nil {
			return err
		}

		raw, ok := strings.CutPrefix(args[0], conf.TagPrefix)
		if !ok {
			return fmt.Errorf("version tag doesn't have prefix: %s", conf.TagPrefix)
		}

		ver, err := semver.Parse(raw)
		if err != nil {
			return err
		}

		v, err := version.Resolve(r, ver, conf.TagPrefix)
		if err != nil {
			return err
		}

		switch tagFlags.Format {
		case "json":
			j, err := json.Marshal(v)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(os.Stdout, string(j))
		case "text":
			_, _ = fmt.Fprintln(os.Stdout, "Tag:    ", v.Tag)
			_, _ = fmt.Fprintln(os.Stdout, "Version:", v.SemVer.String())
			_, _ = fmt.Fprintln(os.Stdout, "Commit: ", v.Commit.FullHash)
		default:
			return fmt.Errorf("unknown format: %s", tagFlags.Format)
		}

		return nil
	},
}

func init() {
	tagCmd.Flags().StringVarP(&tagFlags.Format, "format", "f", "text", "Format to output (json|text)")

	rootCmd.AddCommand(tagCmd)
}
