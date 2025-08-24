//go:build windows

package main

import (
	"os"
	"os/exec"
	"syscall"
)

// osNotifySignals contains the list of notifySignals to handle on Windows.
//
//nolint:gochecknoglobals // Platform-specific signal configuration.
var osNotifySignals = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGABRT,
	syscall.SIGALRM,
	syscall.SIGTERM,
}

// osGetExitStatus returns the exit status on Windows.
func osGetExitStatus(state *os.ProcessState) int {
	return state.ExitCode()
}

// osExecCmd replaces the current process with the command specified by args on Windows.
// Since syscall.Exec is not available on Windows, we use os/exec to run the command
// and then exit with the same exit code.
func osExecCmd(args []string, env []string) error {
	// On Windows, syscall.Exec doesn't exist, so we use os/exec instead
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		return err
	}
	os.Exit(0)
	return nil // This line will never be reached
}

// osChownFrom is a no-op on Windows.
func osChownFrom(string, os.FileInfo) error {
	return nil
}
