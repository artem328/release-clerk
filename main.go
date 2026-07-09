package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/artem328/release-clerk/internal/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type executionResult struct {
	cmd *cobra.Command
	err error
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	done := make(chan executionResult, 1)

	go func() {
		var res executionResult

		res.cmd, res.err = cmd.Execute(ctx)

		done <- res
	}()

	select {
	case <-sig:
		cancel()
	case <-ctx.Done():
	case res := <-done:
		handleResult(res)
		return
	}

	select {
	case <-sig:
		os.Exit(1) // Forcefully shutdown
	case res := <-done:
		handleResult(res)
	}
}

func handleResult(res executionResult) {
	if res.err == nil {
		os.Exit(0)
	}

	var (
		code      int
		showUsage bool
	)

	err := res.err
	reportErr := err

ErrorLoop:
	for err != nil {
		if isCobraSyntaxError(err) {
			reportErr = err
			showUsage = true
			break ErrorLoop
		}

		switch e := err.(type) {
		case cmd.CodeError:
			err = e.Unwrap()
			if code == 0 {
				code = e.Code
			}
			reportErr = err
		case cmd.UsageError:
			showUsage = true
			err = e.Unwrap()
			reportErr = err
		default:
			if err == cmd.ErrSilent {
				reportErr = nil
				break ErrorLoop
			}

			err = errors.Unwrap(err)
		}
	}

	if reportErr != nil {
		_, _ = fmt.Fprintln(os.Stderr, reportErr)
	}

	if showUsage {
		if code == 0 {
			code = 2
		}

		_, _ = fmt.Fprintln(os.Stderr)
		_, _ = fmt.Fprintln(os.Stderr, res.cmd.UsageString())
	}

	if code == 0 {
		code = 1
	}

	os.Exit(code)
}

func isCobraSyntaxError(err error) bool {
	if _, ok := errors.AsType[*pflag.NotExistError](err); ok {
		return true
	}

	if _, ok := errors.AsType[cmd.ArgsError](err); ok {
		return true
	}

	return false
}
