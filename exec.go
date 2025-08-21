package main

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
)

func runCmd(name string, args ...string) (int, error) {
	cmd := exec.CommandContext(context.Background(), name, args...)
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
	signal.Notify(sigc, notifySignals...)
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

	if cmd.ProcessState != nil {
		return getExitStatus(cmd.ProcessState), nil
	}
	return 0, nil
}
