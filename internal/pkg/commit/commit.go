package commit

import (
	"github.com/artem328/release-clerk/git"
	"github.com/artem328/release-clerk/pkg/conventionalcommit"
)

type Commit struct {
	Git          git.Commit
	Conventional conventionalcommit.Commit
}

func FromGitCommit(c git.Commit) Commit {
	return Commit{
		Git:          c,
		Conventional: conventionalcommit.Parse(c.Body),
	}
}

func FromGitCommits(cc []git.Commit) []Commit {
	commits := make([]Commit, 0, len(cc))

	for _, c := range cc {
		commits = append(commits, FromGitCommit(c))
	}

	return commits
}

func FilterOutMergeCommits(commits []Commit) []Commit {
	var i, j int

	for i = 0; i < len(commits); i++ {
		if len(commits[i].Git.Parents) > 1 {
			continue
		}

		commits[j] = commits[i]
		j++
	}

	return commits[:j]
}
