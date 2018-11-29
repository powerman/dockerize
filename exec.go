package main

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func runCmd(name string, args ...string) (int, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	setSysProcAttr(cmd)

	err := cmd.Start()
	if err != nil {
		return 0, err
	}

	sigc := make(chan os.Signal, 8)
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
			_ = cmd.Process.Signal(sig) // pretty sure this doesn't do anything. It seems like the signal is automatically sent to the command?
		}
	}()

	_ = cmd.Wait()

	signal.Stop(sigc)
	close(sigc)

	return cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus(), nil
}
