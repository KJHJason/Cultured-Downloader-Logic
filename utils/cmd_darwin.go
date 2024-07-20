//go:build darwin
// +build darwin

package utils

import (
	"os/exec"
)

func PrepareCmdForBgTask(*exec.Cmd) {
	// do nothing
}
