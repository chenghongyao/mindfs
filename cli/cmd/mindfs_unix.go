//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

func configureBackgroundCommand(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
