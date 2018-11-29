// +build !linux

package main

import (
	"os/exec"
)

func setSysProcAttr(cmd *exec.Cmd) {}
