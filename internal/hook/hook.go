package hook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/artem328/release-clerk/hook"
	"github.com/artem328/release-clerk/internal/config"
	"github.com/artem328/release-clerk/internal/pkg/log"
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

var errHookFailed = errors.New("hook failed")

func RunHooks[I hook.InputPayload, O hook.OutputPayload](
	ctx context.Context,
	hooks []config.Hook,
	i hook.Input[I],
	processor func(hook.Output[O], log.Logger) (hook.Input[I], error),
	l log.Logger,
) (out hook.Output[O], hooked int, err error) {
	// TODO: something is wrong about having logger here
	// should be redesigned
	l = l.Section("hook." + i.Type)
	l.Debugf("Running hooks")

	for j, h := range hooks {
		name := h.Name
		if name == "" {
			name = "hook#" + strconv.Itoa(j)
		}

		hh := Hook{
			Command: h.Command,
			Args:    h.Args,
		}
		hl := l.Section(name)

		hl.Debugf("Starting hook. cmd: %s args: %s", h.Command, h.Args)

		o, err := Run[I, O](ctx, hh, i)
		if errors.Is(err, ErrNotHooked) {
			continue
		}

		out = o

		for _, ll := range out.Logs {
			if ll.Debug {
				hl.Debug(ll.Message)
			} else {
				hl.Log(ll.Message)
			}
		}

		if err != nil {
			hl.Logf("Hook finished with error: %s", err.Error())

			return hook.Output[O]{}, 0, errHookFailed
		}

		i, err = processor(out, l)
		if err != nil {
			hl.Logf("Failed to process hook output: %s", err.Error())

			return hook.Output[O]{}, 0, errHookFailed
		}

		hooked++
		hl.Debug("Finished hook")
	}

	return out, hooked, nil
}
