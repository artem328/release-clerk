package hook

import (
	"github.com/artem328/release-clerk/pkg/semver"
)

const TypePreCommit = "precommit"

type PreCommitInput struct {
	DryRun     bool
	NewVersion semver.Version
	NewTag     string
}

func NewPrecommitInput(p PreCommitInput) Input[PreCommitInput] {
	return Input[PreCommitInput]{
		Type:    TypePreCommit,
		Payload: p,
	}
}

func (PreCommitInput) inputPayloadType() string { return TypePreCommit }

type PreCommitOutput struct{}

func NewPrecommitOutput(logs ...Log) Output[PreCommitOutput] {
	return Output[PreCommitOutput]{
		Type:    TypePreCommit,
		Payload: new(PreCommitOutput),
		Logs:    logs,
	}
}

func NewErrorPrecommitOutput(e Error, logs ...Log) Output[PreCommitOutput] {
	return Output[PreCommitOutput]{
		Type:  TypePreCommit,
		Logs:  logs,
		Error: new(e),
	}
}

func (PreCommitOutput) outputPayloadType() string { return TypePreCommit }
