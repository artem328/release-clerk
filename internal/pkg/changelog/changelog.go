package changelog

import (
	"bytes"
	"time"

	"github.com/artem328/release-clerk/internal/pkg/commit"
	"github.com/artem328/release-clerk/pkg/semver"
)

const DefaultUnmatchedName = "Other"

type Changelog struct {
	Version         semver.Version
	PreviousVersion semver.Version
	Date            time.Time
	Sections        []Section
}

type Section struct {
	Type    string
	Name    string
	Commits []commit.Commit
}

type SectionConfig struct {
	Type   string
	Name   string
	Hidden bool
}

type Config struct {
	Sections                 []SectionConfig
	UnmatchedName            string
	AddBreakingChangeSection bool
	Date                     time.Time
}

func Prepare(ver, prev semver.Version, commits []commit.Commit, conf Config) Changelog {
	sectionsByType := make(map[string]SectionConfig, len(conf.Sections))
	commitsByType := make(map[string][]commit.Commit, len(conf.Sections))
	for _, s := range conf.Sections {
		sectionsByType[s.Type] = s
	}

	const (
		unmatched = "%%unmatched%%"
		breaking  = "%%breaking%%"
	)

	for _, c := range commits {
		if !c.Conventional.IsConventional {
			continue
		}

		if c.Conventional.Type == "chore" && c.Conventional.Scope == "release" {
			// force skip release commits
			continue
		}

		if c.Conventional.IsBreaking && conf.AddBreakingChangeSection {
			commitsByType[breaking] = append(commitsByType[breaking], c)
			continue
		}

		s, ok := sectionsByType[c.Conventional.Type]
		if !ok {
			commitsByType[unmatched] = append(commitsByType[unmatched], c)
			continue
		}

		if s.Hidden {
			continue
		}

		commitsByType[s.Type] = append(commitsByType[s.Type], c)
	}

	changelog := Changelog{
		Version:         ver,
		PreviousVersion: prev,
		Date:            conf.Date,
		Sections:        make([]Section, 0, len(commitsByType)),
	}

	if c := commitsByType[breaking]; len(c) > 0 {
		changelog.Sections = append(changelog.Sections, Section{
			Type:    breaking,
			Name:    "BREAKING CHANGE",
			Commits: c,
		})
	}

	for _, s := range conf.Sections {
		if s.Hidden {
			continue
		}

		c := commitsByType[s.Type]
		if len(c) == 0 {
			continue
		}

		changelog.Sections = append(changelog.Sections, Section{
			Type:    s.Type,
			Name:    s.Name,
			Commits: c,
		})
	}

	if c := commitsByType[unmatched]; len(c) > 0 {
		name := conf.UnmatchedName
		if name == "" {
			name = DefaultUnmatchedName
		}
		changelog.Sections = append(changelog.Sections, Section{
			Type:    unmatched,
			Name:    name,
			Commits: c,
		})
	}

	return changelog
}

func Markdown(changelog Changelog) string {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))

	buf.WriteString("## ")
	buf.WriteString(changelog.Version.String())
	buf.WriteString(" (")
	buf.WriteString(changelog.Date.Format("2006-01-02"))
	buf.WriteString(")")

	for _, s := range changelog.Sections {
		buf.WriteString("\n\n")
		writeMarkdownSection(buf, s)
	}

	return buf.String()
}

func writeMarkdownSection(buf *bytes.Buffer, section Section) {
	buf.WriteString("### ")
	buf.WriteString(section.Name)
	buf.WriteString("\n\n")

	for i, c := range section.Commits {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString("* ")

		if c.Conventional.IsBreaking {
			buf.WriteString("! ")
		}

		if c.Conventional.Scope != "" {
			buf.WriteString("**")
			buf.WriteString(c.Conventional.Scope)
			buf.WriteString("**: ")
		}

		buf.WriteString(c.Conventional.Description)
		buf.WriteString(" (")
		buf.WriteString(c.Git.ShortHash)
		buf.WriteString(")")
	}
}
