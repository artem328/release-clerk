package hook

import (
	"github.com/artem328/release-clerk/pkg/semver"
)

const TypePreCommit = "precommit"

func init() {
	inputPayloadRegistry[TypePreCommit] = func() InputPayload { return &PreCommitInput{} }
	outputPayloadRegistry[TypePreCommit] = func() OutputPayload { return &PreCommitOutput{} }
}

type PreCommitInput struct {
	DryRun     bool
	NewVersion semver.Version
	NewTag     string
}

func NewPrecommitInput(p PreCommitInput) Input {
	return Input{
		Type:    TypePreCommit,
		Payload: new(p),
	}
}

func (PreCommitInput) isInputPayload() {}

type PreCommitOutput struct{}

func NewPrecommitOutput(logs ...Log) Output {
	return Output{
		Type:    TypePreCommit,
		Payload: &PreCommitOutput{},
		Logs:    logs,
	}
}

func NewErrorPrecommitOutput(e Error, logs ...Log) Output {
	return Output{
		Type:  TypePreCommit,
		Logs:  logs,
		Error: new(e),
	}
}

func (PreCommitOutput) isOutputPayload() {}
