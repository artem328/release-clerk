package git

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Commit struct {
	FullHash      string
	ShortHash     string
	AuthorDate    time.Time
	AuthorName    string
	AuthorEmail   string
	CommiterDate  time.Time
	CommiterName  string
	CommiterEmail string
	Body          string
}

const (
	commitFormatSeparator = "%x00"
	commitFormatBoundary  = "%x00%x00%x01"

	commitFormatFullHash      = "%H"
	commitFormatShortHash     = "%h"
	commitFormatAuthorDate    = "%aI"
	commitFormatAuthorName    = "%an"
	commitFormatAuthorEmail   = "%ae"
	commitFormatCommiterDate  = "%cI"
	commitFormatCommiterName  = "%cn"
	commitFormatCommiterEmail = "%ce"
	commitFormatBody          = "%B"
)

const (
	commitBoundaryBytes  = "\x00\x00\x01\n"
	commitSeparatorBytes = "\x00"
)

var commitFormat = [...]string{
	commitFormatFullHash,
	commitFormatShortHash,
	commitFormatAuthorDate,
	commitFormatAuthorName,
	commitFormatAuthorEmail,
	commitFormatCommiterDate,
	commitFormatCommiterName,
	commitFormatCommiterEmail,
	commitFormatBody,
}

var commitFormatString = strings.Join(commitFormat[:], commitFormatSeparator) + commitFormatBoundary

func (r *Repo) GetCommits(pathspec string) ([]Commit, error) {
	if pathspec == "" {
		return r.log()
	}

	return r.log(pathspec)
}

func (r *Repo) GetCommit(ref string) (Commit, error) {
	commits, err := r.log("-1", ref)
	if err != nil {
		return Commit{}, err
	}

	if len(commits) != 1 {
		return Commit{}, fmt.Errorf("unexpected number of commits for %s(%d)", ref, len(commits))
	}

	return commits[0], nil
}

func (r *Repo) StageChanges() error {
	_, err := git(r.ctx, "add", "-A")

	return err
}

func (r *Repo) Commit(msg string) error {
	_, err := git(r.ctx, "commit", "-m", msg)

	return err
}

func (r *Repo) log(args ...string) ([]Commit, error) {
	args = append([]string{"log", "--format=" + commitFormatString}, args...)

	rawCommitsList, err := git(r.ctx, args...)
	if err != nil {
		return nil, err
	}

	rawCommits := bytes.Split(rawCommitsList, []byte(commitBoundaryBytes))
	// remove trailing element which is a product of boundary bytes after the last commit
	rawCommits = rawCommits[:len(rawCommits)-1]
	commits := make([]Commit, 0, len(rawCommits))

	for _, rawCommit := range rawCommits {
		commit, err := r.parseCommit(rawCommit)
		if err != nil {
			return nil, err
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

func (r *Repo) parseCommit(commit []byte) (Commit, error) {
	data := bytes.Split(commit, []byte(commitSeparatorBytes))

	if len(data) != len(commitFormat) {
		return Commit{}, errors.New("commit format mismatch")
	}

	var c Commit

	for i := 0; i < len(commitFormat); i++ {
		switch commitFormat[i] {
		case commitFormatFullHash:
			c.FullHash = string(data[i])
		case commitFormatShortHash:
			c.ShortHash = string(data[i])
		case commitFormatAuthorDate:
			var err error
			c.AuthorDate, err = time.Parse(time.RFC3339, string(data[i]))
			if err != nil {
				return Commit{}, fmt.Errorf("parse author date: %w", err)
			}
		case commitFormatAuthorName:
			c.AuthorName = string(data[i])
		case commitFormatAuthorEmail:
			c.AuthorEmail = string(data[i])
		case commitFormatCommiterDate:
			var err error
			c.CommiterDate, err = time.Parse(time.RFC3339, string(data[i]))
			if err != nil {
				return Commit{}, fmt.Errorf("parse commiter date: %w", err)
			}
		case commitFormatCommiterName:
			c.CommiterName = string(data[i])
		case commitFormatCommiterEmail:
			c.CommiterEmail = string(data[i])
		case commitFormatBody:
			c.Body = string(data[i])
		default:
			return Commit{}, fmt.Errorf("unknown commit data: %v", commitFormat[i])
		}
	}

	return c, nil
}
