package changelog

import (
	"context"

	hookmodel "github.com/artem328/release-clerk/hook"
	"github.com/artem328/release-clerk/internal/config"
	"github.com/artem328/release-clerk/internal/hook"
	"github.com/artem328/release-clerk/internal/pkg/changelog"
	"github.com/artem328/release-clerk/internal/pkg/commit"
	"github.com/artem328/release-clerk/internal/pkg/git"
	"github.com/artem328/release-clerk/internal/pkg/log"
	"github.com/artem328/release-clerk/internal/version"
	"github.com/artem328/release-clerk/pkg/semver"
)

type Config struct {
	Version semver.Version
}

func Generate(ctx context.Context, conf config.Config, genConfig Config) (string, error) {
	repo, err := git.LocateRepo(ctx)
	if err != nil {
		return "", err
	}

	return GenerateForRepo(ctx, repo, conf, genConfig)
}

func GenerateForRepo(ctx context.Context, repo *git.Repo, conf config.Config, genConfig Config) (string, error) {
	var (
		current version.Version
		err     error
	)

	if genConfig.Version.IsZero() {
		current, err = version.Last(repo, conf.TagPrefix)
	} else {
		current, err = version.Resolve(repo, genConfig.Version, conf.TagPrefix)
	}

	if err != nil {
		return "", err
	}

	prev, err := version.Prev(repo, current.SemVer, conf.TagPrefix)
	if err != nil {
		return "", err
	}

	gitCommits, err := repo.GetCommits(git.PathSpec(prev.Commit.FullHash, current.Commit.FullHash))
	if err != nil {
		return "", err
	}

	out, hooked, err := hook.RunHooks(ctx, conf.Hooks, hookmodel.NewCommitTransformInput(hookmodel.CommitTransformInput{
		Commits: gitCommits,
	}), func(o hookmodel.Output[hookmodel.CommitTransformOutput], l log.Logger) (hookmodel.Input[hookmodel.CommitTransformInput], error) {
		return hookmodel.NewCommitTransformInput(hookmodel.CommitTransformInput{
			Commits: o.Payload.Commits,
		}), nil
	}, log.NoLogLogger{})
	if err != nil {
		return "", err
	}

	if hooked > 0 {
		gitCommits = out.Payload.Commits
	}

	commits := commit.FromGitCommits(gitCommits)

	if !conf.IncludeMergeCommits {
		commits = commit.FilterOutMergeCommits(commits)
	}

	sections := make([]changelog.SectionConfig, 0, len(conf.Changelog.Sections))
	for _, s := range conf.Changelog.Sections {
		sections = append(sections, changelog.SectionConfig{
			Type:   s.Type,
			Name:   s.Name,
			Hidden: s.Hidden,
		})
	}

	cl := changelog.Prepare(current.SemVer, prev.SemVer, commits, changelog.Config{
		Sections:                 sections,
		UnmatchedName:            conf.Changelog.UnmatchedSection,
		AddBreakingChangeSection: true,
		Date:                     current.Commit.CommiterDate,
	})

	return changelog.Markdown(cl), nil
}
