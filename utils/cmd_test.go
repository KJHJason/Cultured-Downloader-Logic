package utils

import (
	"os/exec"
	"testing"
	"time"
)

func TestInterruptProcess(t *testing.T) {
	pythonScript := "import time\nwhile True:\n    print('hello')\n    time.sleep(1)\n"
	cmd := exec.Command("python", "-c", pythonScript)
	err := cmd.Start()
	t.Logf("Started process with PID %d", cmd.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)
	err = InterruptProcess(cmd)
	if err != nil {
		t.Fatal(err)
	}

	err = cmd.Wait()
	if err != nil {
		t.Fatal(err)
	}
}
