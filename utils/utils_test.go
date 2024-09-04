package utils

import (
	"testing"
)

func TestGetUnusedTcpPort(t *testing.T) {
	port, err := GetUnusedTcpPort()
	if err != nil {
		t.Errorf("Error getting unused port: %v", err)
	}
	t.Logf("Unused port: %d", port)
}
