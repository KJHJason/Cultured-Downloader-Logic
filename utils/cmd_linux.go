//go:build linux
// +build linux

package utils

import (
	"os/exec"
)

func PrepareCmdForBgTask(*exec.Cmd) {
	// do nothing
}
