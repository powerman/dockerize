//go:build unix

package main

import (
	"os"
	"os/exec"
	"syscall"
)

// osNotifySignals contains the list of osNotifySignals to handle on Unix systems.
//
//nolint:gochecknoglobals // Platform-specific signal configuration.
var osNotifySignals = []os.Signal{
	syscall.SIGHUP,
	syscall.SIGINT,
	syscall.SIGQUIT,
	syscall.SIGABRT,
	syscall.SIGUSR1,
	syscall.SIGUSR2,
	syscall.SIGALRM,
	syscall.SIGTERM,
}

// osGetExitStatus returns the exit status from ProcessState on Unix systems.
func osGetExitStatus(state *os.ProcessState) int {
	return state.Sys().(syscall.WaitStatus).ExitStatus()
}

// osExecCmd replaces the current process with the command specified by args on Unix systems.
func osExecCmd(args []string, env []string) error {
	arg0, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	return syscall.Exec(arg0, args, env) //nolint:gosec // False positive.
}

// osChownFrom works like `chown --reference` on Unix systems.
// It changes the ownership of dst to match the reference file info.
// If ownership change is not permitted, it ignores the error.
func osChownFrom(dst string, reference os.FileInfo) error {
	refStat, ok := reference.Sys().(*syscall.Stat_t)
	if ok {
		err := os.Chown(dst, int(refStat.Uid), int(refStat.Gid))
		if os.IsPermission(err) {
			err = nil
		}
		return err
	}
	return nil
}
