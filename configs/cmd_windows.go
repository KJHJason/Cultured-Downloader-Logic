// +build windows

package configs

import (
	"os/exec"
	"syscall"
)

func PrepareCmdForBgTask(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}
