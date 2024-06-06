//go:build darwin
// +build darwin

package utils

import (
	"os/exec"
)

func PrepareCmdForBgTask(cmd *exec.Cmd) {
	// do nothing
}
