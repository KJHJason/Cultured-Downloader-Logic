package cache

import (
	"encoding/binary"
)

func ParseInt64(value int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	return buf
}

func ParseInt(value int) []byte {
	return ParseInt64(int64(value))
}

func ParseBytesToInt64(value []byte) int64 {
	return int64(binary.LittleEndian.Uint64(value))
}

func ParseBytesToInt(value []byte) int {
	return int(ParseBytesToInt64(value))
}
