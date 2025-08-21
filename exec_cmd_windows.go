//go:build windows

package main

import (
	"os"
	"os/exec"
)

func execCmd(args []string, env []string) error {
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
