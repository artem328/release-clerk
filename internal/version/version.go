package version

import (
	"errors"
	"slices"
	"strings"

	"github.com/artem328/release-clerk/internal/pkg/git"
	"github.com/artem328/release-clerk/pkg/semver"
)

type Version struct {
	Tag    string
	SemVer semver.Version
	Commit git.Commit
}

var ErrNoVersion = errors.New("no version")

func Last(r *git.Repo, tagPrefix string) (Version, error) {
	vv, err := versions(r, tagPrefix)
	if err != nil {
		return Version{}, err
	}

	if len(vv) == 0 {
		return Version{}, nil
	}

	tag, c, err := tagAndCommit(r, vv[0], tagPrefix)
	if err != nil {
		return Version{}, err
	}

	return Version{
		Tag:    tag,
		SemVer: vv[0],
		Commit: c,
	}, nil
}

func Resolve(r *git.Repo, ver semver.Version, tagPrefix string) (Version, error) {
	vv, err := versions(r, tagPrefix)
	if err != nil {
		return Version{}, err
	}

	if len(vv) == 0 {
		return Version{}, ErrNoVersion
	}

	i, ok := slices.BinarySearchFunc(vv, ver, func(v, target semver.Version) int {
		return target.Compare(v)
	})
	if !ok {
		return Version{}, ErrNoVersion
	}

	tag, c, err := tagAndCommit(r, vv[i], tagPrefix)
	if err != nil {
		return Version{}, err
	}

	return Version{
		Tag:    tag,
		SemVer: vv[i],
		Commit: c,
	}, nil
}

func Prev(r *git.Repo, ver semver.Version, tagPrefix string) (Version, error) {
	vv, err := versions(r, tagPrefix)
	if err != nil {
		return Version{}, err
	}

	if len(vv) == 0 {
		return Version{}, nil
	}

	i, ok := slices.BinarySearchFunc(vv, ver, func(v, target semver.Version) int {
		return target.Compare(v)
	})
	if !ok {
		return Version{}, ErrNoVersion
	}

	if i == len(vv)-1 {
		return Version{}, nil
	}

	tag, c, err := tagAndCommit(r, vv[i+1], tagPrefix)
	if err != nil {
		return Version{}, err
	}

	return Version{
		Tag:    tag,
		SemVer: vv[i+1],
		Commit: c,
	}, nil
}

func tagAndCommit(r *git.Repo, ver semver.Version, tagPrefix string) (string, git.Commit, error) {
	tag := tagPrefix + ver.String()

	c, err := r.GetCommit(tag)
	if err != nil {
		return "", git.Commit{}, err
	}

	return tag, c, nil
}

func versions(r *git.Repo, tagPrefix string) ([]semver.Version, error) {
	tags, err := r.Tags()
	if err != nil {
		return nil, err
	}

	semvers := make([]semver.Version, 0, len(tags))
	for _, tag := range tags {
		t, ok := strings.CutPrefix(tag, tagPrefix)
		if !ok {
			continue
		}

		v, err := semver.Parse(t)
		if err != nil {
			continue
		}

		semvers = append(semvers, v)
	}

	slices.SortStableFunc(semvers, func(a, b semver.Version) int {
		return b.Compare(a)
	})

	return semvers, nil
}
