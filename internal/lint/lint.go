package lint

import (
	"context"
	"fmt"

	hookmodel "github.com/artem328/release-clerk/hook"
	"github.com/artem328/release-clerk/internal/config"
	"github.com/artem328/release-clerk/internal/hook"
	"github.com/artem328/release-clerk/internal/pkg/commit"
	"github.com/artem328/release-clerk/internal/pkg/git"
	"github.com/artem328/release-clerk/internal/pkg/log"
)

type IssueKind string

const (
	IssueWarning IssueKind = "warning"
	IssueError   IssueKind = "error"
)

type Issue struct {
	Kind    IssueKind
	Commit  commit.Commit
	Message string
}

type Config struct {
	ErrorOnUnknownType bool
}

func Run(ctx context.Context, conf config.Config, lintConf Config, pathspec string) ([]Issue, error) {
	r, err := git.LocateRepo(ctx)
	if err != nil {
		return nil, err
	}

	gitCommits, err := r.GetCommits(pathspec)
	if err != nil {
		return nil, err
	}

	out, hooked, err := hook.RunHooks(ctx, conf.Hooks, hookmodel.NewCommitTransformInput(hookmodel.CommitTransformInput{
		Commits: gitCommits,
	}), func(o hookmodel.Output[hookmodel.CommitTransformOutput], l log.Logger) (hookmodel.Input[hookmodel.CommitTransformInput], error) {
		return hookmodel.NewCommitTransformInput(hookmodel.CommitTransformInput{
			Commits: o.Payload.Commits,
		}), nil
	}, log.NoLogLogger{})
	if err != nil {
		return nil, err
	}

	if hooked > 0 {
		gitCommits = out.Payload.Commits
	}

	commits := commit.FromGitCommits(gitCommits)

	if !conf.IncludeMergeCommits {
		commits = commit.FilterOutMergeCommits(commits)
	}

	allowedTypes := make(map[string]struct{}, len(conf.Rules))
	for _, r := range conf.Rules {
		allowedTypes[r.Type] = struct{}{}
	}

	var issues []Issue

	unknownTypeKind := IssueWarning
	if lintConf.ErrorOnUnknownType {
		unknownTypeKind = IssueError
	}

	for _, c := range commits {
		if !c.Conventional.IsConventional {
			issues = append(issues, Issue{
				Kind:    IssueError,
				Commit:  c,
				Message: "The commit does not follow conventional commit format",
			})
			continue
		}

		if _, ok := allowedTypes[c.Conventional.Type]; !ok {
			issues = append(issues, Issue{
				Kind:    unknownTypeKind,
				Commit:  c,
				Message: fmt.Sprintf("Type %s is not defined in config", c.Conventional.Type),
			})
		}
	}

	return issues, nil
}
