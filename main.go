package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/artem328/release-clerk/cmd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	done := make(chan error, 1)

	go func() {
		done <- cmd.Execute(ctx)
	}()

	select {
	case <-sig:
		cancel()
	case <-ctx.Done():
	case err := <-done:
		if err != nil {
			os.Exit(1)
		}
		return
	}

	select {
	case <-sig:
		os.Exit(1) // Forcefully shutdown
	case err := <-done:
		if err != nil {
			os.Exit(1)
		}
	}
}
