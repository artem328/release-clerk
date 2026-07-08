package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/artem328/release-clerk/git"
	"github.com/artem328/release-clerk/hook"
)

func main() {
	var isError bool

	flag.BoolVar(&isError, "error", false, "return an error")
	flag.Parse()

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	var h hook.Input[hook.CommitTransformInput]
	if err := json.Unmarshal(data, &h); errors.Is(err, hook.ErrInputTypeMismatch) {
		return
	} else if err != nil {
		panic(err)
	}

	commits := make([]git.Commit, 0, len(h.Payload.Commits)*2)
	for _, commit := range h.Payload.Commits {
		dup := commit

		lines := strings.Split(commit.Body, "\n")
		lines[0] += " (duplicated)"
		dup.Body = strings.Join(lines, "\n")
		commits = append(commits, dup, commit)
	}

	logs := []hook.Log{{Message: "Log1"}, {Message: "Log2"}, {Message: "Log3", Debug: true}}

	var out hook.Output[hook.CommitTransformOutput]
	if isError {
		out = hook.NewErrorCommitTransformOutput(hook.Error{Message: "Hook failed"}, logs...)
	} else {
		out = hook.NewCommitTransformOutput(hook.CommitTransformOutput{
			Commits: commits,
		}, logs...)
	}

	j, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}

	_, _ = fmt.Fprint(os.Stdout, string(j))
}
