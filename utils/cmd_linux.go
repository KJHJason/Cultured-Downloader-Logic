//go:build linux
// +build linux

package utils

import (
	"os/exec"
	"syscall"
)

func PrepareCmdForBgTask(cmd *exec.Cmd) {
	// do nothing
}

func prepareCmdForNewWindow(cmd *exec.Cmd) {
	// do nothing
}

func InterruptProcess(cmd *exec.Cmd) error {
	return cmd.Process.Signal(syscall.SIGINT)
}
