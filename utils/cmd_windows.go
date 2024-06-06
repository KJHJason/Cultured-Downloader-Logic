//go:build windows
// +build windows

package utils

import (
	"os/exec"
	"syscall"
)

func PrepareCmdForBgTask(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
