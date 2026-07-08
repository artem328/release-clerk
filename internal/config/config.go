package config

import "fmt"

type Bump uint8

const (
	BumpNone Bump = iota
	BumpPatch
	BumpMinor
	BumpMajor
)

func (b *Bump) UnmarshalText(text []byte) error {
	switch string(text) {
	case "none":
		*b = BumpNone
	case "patch":
		*b = BumpPatch
	case "minor":
		*b = BumpMinor
	case "major":
		*b = BumpMajor
	default:
		return fmt.Errorf("unknown Bump type: %q", string(text))
	}

	return nil
}

type Branch struct {
	Name string `yaml:"name"`
}

type Rule struct {
	Type  string `yaml:"type"`
	Scope string `yaml:"scope"`
	Bump  Bump   `yaml:"bump"`
}

type ChangelogSection struct {
	Type   string `yaml:"type"`
	Name   string `yaml:"name"`
	Hidden bool   `yaml:"hidden"`
}

type Changelog struct {
	Path             string             `yaml:"path"`
	UnmatchedSection string             `yaml:"unmatchedSection"`
	Sections         []ChangelogSection `yaml:"sections"`
}

type Hook struct {
	Name    string   `yaml:"name"`
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
}

type Config struct {
	TagPrefix           string    `yaml:"tagPrefix"`
	DisableMajor        bool      `yaml:"disableMajor"`
	DisablePush         bool      `yaml:"disablePush"`
	IncludeMergeCommits bool      `yaml:"includeMergeCommits"`
	Branches            []Branch  `yaml:"branches"`
	Rules               []Rule    `yaml:"rules"`
	Changelog           Changelog `yaml:"changelog"`
	Hooks               []Hook    `yaml:"hooks"`
}
