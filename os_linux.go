//go:build linux

package main

import (
	"os/exec"
	"syscall"
)

func osSetSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
}
