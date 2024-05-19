package cache

import (
	"encoding/binary"
	"time"
)

func ParseInt64(value int64) []byte {
	buf := make([]byte, 8)
	binary.NativeEndian.PutUint64(buf, uint64(value))
	return buf
}

func ParseInt(value int) []byte {
	return ParseInt64(int64(value))
}

func ParseBytesToInt64(value []byte) int64 {
	if len(value) != 8 {
		return -1
	}
	return int64(binary.NativeEndian.Uint64(value))
}

func ParseBytesToInt(value []byte) int {
	return int(ParseBytesToInt64(value))
}

func ParseDateTimeToBytes(datetime time.Time) []byte {
	sec := datetime.Unix()
	nSec := datetime.Nanosecond()

	buf := make([]byte, 16)
	binary.NativeEndian.PutUint64(buf, uint64(sec))
	binary.NativeEndian.PutUint64(buf[8:], uint64(nSec))
	return buf
}

func ParseBytesToDateTime(value []byte) time.Time {
	if len(value) != 16 {
		return time.Time{}
	}
	sec := int64(binary.NativeEndian.Uint64(value))
	nSec := int64(binary.NativeEndian.Uint64(value[8:]))
	return time.Unix(sec, nSec)
}
