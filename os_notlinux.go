//go:build !linux

package main

import (
	"os/exec"
)

func osSetSysProcAttr(cmd *exec.Cmd) {}
