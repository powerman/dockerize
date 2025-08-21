//go:build windows

package main

import (
	"os"
	"syscall"
)

// notifySignals contains the list of notifySignals to handle on Windows.
//
//nolint:gochecknoglobals // Platform-specific signal configuration.
var notifySignals = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGABRT,
	syscall.SIGALRM,
	syscall.SIGTERM,
}

// getExitStatus returns the exit status on Windows.
func getExitStatus(state *os.ProcessState) int {
	return state.ExitCode()
}
