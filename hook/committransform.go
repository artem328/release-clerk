package hook

import (
	"github.com/artem328/release-clerk/git"
)

const TypeCommitTransform = "commit_transform"

type CommitTransformInput struct {
	Commits []git.Commit
}

func NewCommitTransformInput(i CommitTransformInput) Input[CommitTransformInput] {
	return Input[CommitTransformInput]{
		Type:    TypeCommitTransform,
		Payload: i,
	}
}

func (CommitTransformInput) inputPayloadType() string { return TypeCommitTransform }

type CommitTransformOutput struct {
	Commits []git.Commit
}

func NewCommitTransformOutput(o CommitTransformOutput, logs ...Log) Output[CommitTransformOutput] {
	return Output[CommitTransformOutput]{
		Type:    TypeCommitTransform,
		Logs:    logs,
		Payload: new(o),
	}
}

func NewErrorCommitTransformOutput(e Error, logs ...Log) Output[CommitTransformOutput] {
	return Output[CommitTransformOutput]{
		Type:  TypeCommitTransform,
		Logs:  logs,
		Error: new(e),
	}
}

func (CommitTransformOutput) outputPayloadType() string { return TypeCommitTransform }
