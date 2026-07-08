package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type Repo struct {
	ctx  context.Context
	root string
}

func LocateRepo(ctx context.Context) (*Repo, error) {
	root, err := runGit(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}

	return &Repo{ctx: ctx, root: string(root)}, nil
}

func (r *Repo) Root() string {
	return r.root
}

func (r *Repo) Tags() ([]string, error) {
	tagsList, err := runGit(r.ctx, "tag", "--list")
	if err != nil {
		return nil, nil
	}

	rawTags := bytes.Split(tagsList, []byte("\n"))
	tags := make([]string, 0, len(rawTags))
	for _, t := range rawTags {
		tag := strings.TrimSpace(string(t))
		if tag != "" {
			tags = append(tags, tag)
		}
	}

	return tags, nil
}

func (r *Repo) CurrentBranch() (string, error) {
	branch, err := runGit(r.ctx, "branch", "--show-current")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(branch)), nil
}

func (r *Repo) AddTag(tag, annotation string) error {
	args := make([]string, 0, 5)
	args = append(args, "tag")

	if annotation != "" {
		args = append(args, "-a", tag, "-m", annotation)
	} else {
		args = append(args, tag)
	}

	_, err := runGit(r.ctx, args...)

	return err
}

func (r *Repo) Push(branch string) error {
	_, err := runGit(r.ctx, "push", "origin", branch, "--follow-tags")

	return err
}

func runGit(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	stdout := bytes.NewBuffer(make([]byte, 0, 1024))
	stderr := bytes.NewBuffer(make([]byte, 0, 1024))

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w\n%s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
