package main

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func runCmd(name string, args ...string) (int, error) {
	cmd := exec.Command(name, args...) //nolint:gosec // Subprocess launched with variable.
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	setSysProcAttr(cmd)

	err := cmd.Start()
	if err != nil {
		return 0, err
	}

	const sigcSize = 8
	sigc := make(chan os.Signal, sigcSize)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGABRT,
		syscall.SIGUSR1,
		syscall.SIGUSR2,
		syscall.SIGALRM,
		syscall.SIGTERM,
	)
	go func() {
		for sig := range sigc {
			// This will duplicate some signals if they're
			// sent to all processes in current group (like
			// when Ctrl-C is pressed in shell).
			_ = cmd.Process.Signal(sig)
		}
	}()

	_ = cmd.Wait()

	signal.Stop(sigc)
	close(sigc)

	return cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus(), nil
}
