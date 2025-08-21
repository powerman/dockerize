//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

func execCmd(args []string, env []string) error {
	arg0, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	return syscall.Exec(arg0, args, env) //nolint:gosec // False positive.
}
