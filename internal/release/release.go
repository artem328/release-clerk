package release

import (
	"bytes"
	"context"
	"errors"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	hookmodel "github.com/artem328/release-clerk/hook"
	"github.com/artem328/release-clerk/internal/config"
	"github.com/artem328/release-clerk/internal/hook"
	"github.com/artem328/release-clerk/internal/pkg/changelog"
	"github.com/artem328/release-clerk/internal/pkg/commit"
	"github.com/artem328/release-clerk/internal/pkg/git"
	"github.com/artem328/release-clerk/internal/pkg/log"
	"github.com/artem328/release-clerk/internal/version"
	"github.com/artem328/release-clerk/pkg/conventionalcommit"
	"github.com/artem328/release-clerk/pkg/semver"
)

type Release struct {
	Tag     string
	Version version.Version
}

type rule struct {
	Type  string
	Scope string
	Bump  config.Bump
}

func newRule(r config.Rule) rule {
	return rule{
		Type:  r.Type,
		Scope: r.Scope,
		Bump:  r.Bump,
	}
}

func (r rule) Match(c conventionalcommit.Commit) bool {
	hasType := r.Type != ""
	hasScope := r.Scope != ""
	matchType := strings.EqualFold(r.Type, c.Type)
	matchScope := strings.EqualFold(r.Scope, c.Scope)

	if hasType && hasScope {
		return matchType && matchScope
	}

	return (hasType && matchType) || (hasScope && matchScope)
}

func (r rule) Compare(other rule) int {
	// 1. type & scope
	// 2. type | scope

	ruleHasTypeAndScope := r.Type != "" && r.Scope != ""
	otherHasTypeAndScope := other.Type != "" && other.Scope != ""

	if (ruleHasTypeAndScope && otherHasTypeAndScope) || (!ruleHasTypeAndScope && !otherHasTypeAndScope) {
		return 0
	}

	if ruleHasTypeAndScope {
		return -1
	}

	return 1
}

func Run(ctx context.Context, conf config.Config, dryRun bool, l log.Logger) (*Release, error) {
	if dryRun {
		l.Log("╔════════════════════════════════════════╗")
		l.Log("║                DRY RUN                 ║")
		l.Log("╚════════════════════════════════════════╝")
	}

	l = l.Section("release")

	repo, err := git.LocateRepo(ctx)
	if err != nil {
		return nil, err
	}

	branch, err := repo.CurrentBranch()
	if err != nil {
		return nil, err
	}

	var onAValidBranch bool
	for _, b := range conf.Branches {
		if b.Name == branch {
			onAValidBranch = true
			break
		}
	}

	if !onAValidBranch {
		l.Logf("Branch `%s` is not targeted. Doing nothing", branch)
		return nil, nil
	}

	lastVersion, err := version.Last(repo, conf.TagPrefix)
	if err != nil {
		return nil, err
	}

	vlog := l.Section("last-version")
	if !lastVersion.SemVer.IsZero() {
		vlog.Log("Version found")
		vlog.Logf("      Tag: %s", lastVersion.Tag)
		vlog.Logf("  Version: %s", lastVersion.SemVer.String())
	} else {
		vlog.Log("No versions found")
		vlog.Logf("  Version: %s", lastVersion.SemVer.String())
	}

	gitCommits, err := repo.GetCommits(git.PathSpec(lastVersion.Commit.FullHash, "HEAD"))
	if err != nil {
		return nil, err
	}

	commits := commit.FromGitCommits(gitCommits)

	clog := l.Section("commits")
	if len(commits) > 0 {
		commitNoun := "commit"
		if len(commits) > 1 {
			commitNoun = "commits"
		}
		clog.Logf("Found %d %s since last version", len(commits), commitNoun)
	} else {
		clog.Log("No commits found since last version. Doing nothing")
		return nil, nil
	}

	rules := make([]rule, 0, len(conf.Rules))
	for _, r := range conf.Rules {
		rules = append(rules, newRule(r))
	}

	slices.SortStableFunc(rules, func(a, b rule) int {
		return a.Compare(b)
	})

	nvlog := l.Section("new-version")
	newVersion, bump := bumpVersion(lastVersion.SemVer, commits, rules, conf.DisableMajor)
	newTag := conf.TagPrefix + newVersion.String()

	if bump {
		nvlog.Log("New version detected")
		nvlog.Logf("      Tag: %s", newTag)
		nvlog.Logf("  Version: %s", newVersion.String())
	} else {
		nvlog.Log("No changes detected for new version. Doing nothing")
		return nil, nil
	}

	cllog := l.Section("changelog")
	sections := make([]changelog.SectionConfig, 0, len(conf.Changelog.Sections))
	for _, s := range conf.Changelog.Sections {
		sections = append(sections, changelog.SectionConfig{
			Type:   s.Type,
			Name:   s.Name,
			Hidden: s.Hidden,
		})
	}

	cl := changelog.Prepare(newVersion, lastVersion.SemVer, commits, changelog.Config{
		Sections:                 sections,
		UnmatchedName:            conf.Changelog.UnmatchedSection,
		AddBreakingChangeSection: true,
		Date:                     time.Now(),
	})
	clData := changelog.Markdown(cl)
	cllog.Debug("\n", clData)
	filename := conf.Changelog.Path
	if filename == "" {
		filename = "CHANGELOG.md"
	}
	if !dryRun {
		if err := writeChangelog(filename, clData); err != nil {
			return nil, err
		}
		cllog.Logf("Changelog written to %s", filename)
	} else {
		cllog.Logf("Changelog write to %s is skipped due to dry-run", filename)
	}

	precommit := hookmodel.NewPrecommitInput(hookmodel.PreCommitInput{
		DryRun:     dryRun,
		NewVersion: newVersion,
		NewTag:     newTag,
	})

	if len(conf.Hooks) > 0 {
		_, _, err := runHooks(ctx, conf.Hooks, precommit, func(o hookmodel.Output, l log.Logger) hookmodel.Input { return precommit }, l)
		if err != nil {
			return nil, err
		}
	}

	commlog := l.Section("commit")
	if !dryRun {
		if err := repo.StageChanges(); err != nil {
			return nil, err
		}
		if err := repo.Commit("chore(release): version " + newVersion.String()); err != nil {
			return nil, err
		}
		commlog.Log("Changes commited")
	} else {
		commlog.Log("Skipping commit due to dry-run")
	}

	tlog := l.Section("tag")
	if !dryRun {
		if err := repo.AddTag(newTag, "Release "+newVersion.String()); err != nil {
			return nil, err
		}
		tlog.Log("New version tag is added")
	} else {
		tlog.Log("Skipping tagging due to dry-run")
	}

	plog := l.Section("push")
	if !dryRun {
		if err := repo.Push(branch); err != nil {
			return nil, err
		}
		plog.Log("Changes pushed")
	} else {
		plog.Log("Skipping push due to dry-run")
	}

	if dryRun {
		return nil, nil
	}

	v, err := version.Resolve(repo, newVersion, conf.TagPrefix)
	if err != nil {
		return nil, err
	}

	return &Release{
		Tag:     newTag,
		Version: v,
	}, nil
}

func bumpVersion(version semver.Version, commits []commit.Commit, rules []rule, disableMajor bool) (semver.Version, bool) {
	var bump config.Bump

Commits:
	for _, c := range commits {
		if !c.Conventional.IsConventional {
			continue
		}

		if c.Conventional.IsBreaking {
			bump = config.BumpMajor
			break
		}

		if c.Conventional.Type == "chore" && c.Conventional.Scope == "release" {
			// skip commits created by the clerk
			continue
		}

		for _, r := range rules {
			if r.Match(c.Conventional) && r.Bump > bump {
				bump = r.Bump

				if bump == config.BumpMajor {
					break Commits
				}

				break
			}
		}
	}

	switch bump {
	case config.BumpNone:
		return version, false
	case config.BumpPatch:
		return version.BumpPatch(), true
	case config.BumpMinor:
		return version.BumpMinor(), true
	case config.BumpMajor:
		if disableMajor {
			return version.BumpMinor(), true
		}
		return version.BumpMajor(), true
	default:
		panic("unexpected bump type. this is a bug")
	}
}

func writeChangelog(file string, changelog string) error {
	const title = "# CHANGELOG"

	current, err := os.ReadFile(file)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if len(current) == 0 {
		if err := os.WriteFile(file, []byte(title+"\n\n"+changelog), 0644); err != nil {
			return err
		}

		return nil
	}

	data := make([]byte, 0, len(current)+len(changelog)+len(title)+2)

	if !bytes.HasPrefix(current, []byte(title)) {
		data = append(data, []byte(title)...)
		data = append(data, "\n\n"...)
	} else {
		for i := len(title); i < len(current)-5; i++ {
			// find first occurrence of h2
			if current[i] == '\n' &&
				current[i+1] == '\n' &&
				current[i+2] == '#' &&
				current[i+3] == '#' &&
				current[i+4] == ' ' {
				data = append(data, current[:i]...)
				data = append(data, "\n\n"...)
				current = current[i+2:]
				break
			}
		}
	}

	data = append(data, changelog...)
	data = append(data, "\n\n"...)
	data = append(data, current...)

	return os.WriteFile(file, data, 0644)
}

var errHookFailed = errors.New("hook failed")

func runHooks(
	ctx context.Context,
	hooks []config.Hook,
	i hookmodel.Input,
	processor func(hookmodel.Output, log.Logger) hookmodel.Input,
	l log.Logger,
) (out hookmodel.Output, hooked int, err error) {
	l = l.Section("hook." + i.Type)
	l.Log("Running hooks")

	for j, h := range hooks {
		name := h.Name
		if name == "" {
			name = "hook#" + strconv.Itoa(j)
		}

		hh := hook.Hook{
			Command: h.Command,
			Args:    h.Args,
		}
		hl := l.Section(name)

		hl.Debugf("Starting hook. cmd: %s args: %s", h.Command, h.Args)

		out, err = hook.Run(ctx, hh, i)
		if errors.Is(err, hook.ErrNotHooked) {
			continue
		}

		for _, ll := range out.Logs {
			if ll.Debug {
				hl.Debug(ll.Message)
			} else {
				hl.Log(ll.Message)
			}
		}

		if err != nil {
			hl.Logf("Hook finished with error: %s", err.Error())

			return hookmodel.Output{}, 0, errHookFailed
		}

		i = processor(out, l)

		hooked++
		hl.Debug("Finished hook")
	}

	return out, hooked, nil
}
