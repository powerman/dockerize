//go:build !windows

package main

import (
	"os"
	"syscall"
)

// notifySignals contains the list of notifySignals to handle on Unix systems.
//
//nolint:gochecknoglobals // Platform-specific signal configuration.
var notifySignals = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGABRT,
	syscall.SIGUSR1,
	syscall.SIGUSR2,
	syscall.SIGALRM,
	syscall.SIGTERM,
}

// getExitStatus returns the exit status from ProcessState on Unix systems.
func getExitStatus(state *os.ProcessState) int {
	return state.Sys().(syscall.WaitStatus).ExitStatus()
}
