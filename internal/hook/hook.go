package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/artem328/release-clerk/hook"
)

var ErrNotHooked = errors.New("not hooked")

type Hook struct {
	Command string
	Args    []string
}

func Run[I hook.InputPayload, O hook.OutputPayload](ctx context.Context, h Hook, i hook.Input[I]) (hook.Output[O], error) {
	cmd := exec.CommandContext(ctx, h.Command, h.Args...)
	stderr := bytes.NewBuffer(make([]byte, 0, 1024))
	stdout := bytes.NewBuffer(make([]byte, 0, 1024))

	j, err := json.Marshal(i)
	if err != nil {
		return hook.Output[O]{}, err
	}

	cmd.Stdin = bytes.NewBuffer(j)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return hook.Output[O]{}, err
	}

	if stdout.Len() == 0 {
		return hook.Output[O]{}, ErrNotHooked
	}

	var o hook.Output[O]

	if err := json.Unmarshal(stdout.Bytes(), &o); err != nil {
		return hook.Output[O]{}, err
	}

	if o.Type != i.Type {
		return hook.Output[O]{}, fmt.Errorf("hook responded with incorrect type %s, while expected %s", o.Type, i.Type)
	}

	return o, o.Err()
}
