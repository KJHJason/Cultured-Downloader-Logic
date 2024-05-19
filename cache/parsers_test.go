package cache

import (
	"testing"
	"time"
)

func TestDateTimeParser(t *testing.T) {
	// Test parsing a datetime to bytes
	datetime := time.Now()
	bytes := ParseDateTimeToBytes(datetime)
	if len(bytes) != 16 {
		t.Errorf("Expected 8 bytes, got %d", len(bytes))
	}

	// Test parsing bytes to a datetime
	newDatetime := ParseBytesToDateTime(bytes)
	if !datetime.Equal(newDatetime) {
		t.Errorf("Expected %v, got %v", datetime, newDatetime)
	}
}

func TestIntParser(t *testing.T) {
	// Test parsing an int to bytes
	value := 123
	bytes := ParseInt(value)
	if len(bytes) != 8 {
		t.Errorf("Expected 8 bytes, got %d", len(bytes))
	}

	// Test parsing bytes to an int
	newValue := ParseBytesToInt(bytes)
	if value != newValue {
		t.Errorf("Expected %d, got %d", value, newValue)
	}
}
