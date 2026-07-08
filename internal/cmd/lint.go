package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/artem328/release-clerk/internal/lint"
	"github.com/spf13/cobra"
)

var lintFlags = struct {
	ErrorOnUnknownType bool
	Format             string
}{}

var lintCmd = &cobra.Command{
	Use:       "lint <pathspec>",
	Short:     "Lint commits within pathspec range",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"pathspec"},
	RunE: func(cmd *cobra.Command, args []string) error {
		conf, err := loadConfig()
		if err != nil {
			return err
		}

		issues, err := lint.Run(
			cmd.Context(),
			conf,
			lint.Config{ErrorOnUnknownType: lintFlags.ErrorOnUnknownType},
			args[0],
		)
		if err != nil {
			return err
		}

		var hasError bool

		switch lintFlags.Format {
		case "text":
			for _, i := range issues {
				level := "warn"

				switch i.Kind {
				case lint.IssueError:
					hasError = true
					level = "erro"
				default:
				}

				_, _ = fmt.Fprintf(os.Stdout, "[%s][%s] %s\n", level, i.Commit.Git.ShortHash, i.Message)
			}
		case "json":
			for _, i := range issues {
				if i.Kind == lint.IssueError {
					hasError = true
					break
				}
			}

			j, err := json.Marshal(issues)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(os.Stdout, string(j))
		default:
			return fmt.Errorf("unknown format: %s", lintFlags.Format)
		}

		if hasError {
			return ErrSilent
		}

		return nil
	},
}

func init() {
	lintCmd.Flags().BoolVarP(
		&lintFlags.ErrorOnUnknownType,
		"unknown-type-error",
		"t",
		false,
		"Fail on unknown type",
	)
	lintCmd.Flags().StringVarP(
		&lintFlags.Format,
		"format",
		"f",
		"text",
		"Format to output (json|text)",
	)

	rootCmd.AddCommand(lintCmd)
}
