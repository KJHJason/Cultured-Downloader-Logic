//go:build windows
// +build windows

package utils

import (
	"os/exec"
	"syscall"

	// "golang.org/x/sys/windows"
)

func PrepareCmdForBgTask(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
}

func prepareCmdForNewWindow(cmd *exec.Cmd) {
	const CREATE_NEW_CONSOLE = 0x10
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: CREATE_NEW_CONSOLE,
		NoInheritHandles: true,
	}
}

func InterruptProcess(cmd *exec.Cmd) error {
	// TODO: find a way to interrupt the process
	cmd.Process.Signal(syscall.SIGTERM)
	// // since cmd.Process.Signal() is not supported
	// // on Windows, we have to use the Windows API.
	// dll, err := windows.LoadDLL("kernel32.dll")
	// if err != nil {
	// 	return err
	// }
	// defer dll.Release()

	// // https://github.com/mattn/goreman/blob/e9150e84f13c37dff0a79b8faed5b86522f3eb8e/proc_windows.go#L16-L51
	// pid := uintptr(cmd.Process.Pid)

	// f, err := dll.FindProc("AttachConsole")
	// if err != nil {
	// 	return err
	// }
	// r1, _, err := f.Call(pid)
	// if r1 == 0 && err != syscall.ERROR_ACCESS_DENIED {
	// 	return err
	// }

	// f, err = dll.FindProc("SetConsoleCtrlHandler")
	// if err != nil {
	// 	return err
	// }
	// r1, _, err = f.Call(0, 1)
	// if r1 == 0 {
	// 	return err
	// }
	// f, err = dll.FindProc("GenerateConsoleCtrlEvent")
	// if err != nil {
	// 	return err
	// }

	// r1, _, err = f.Call(windows.CTRL_BREAK_EVENT, pid)
	// if r1 == 0 {
	// 	return err
	// }
	return nil
}
