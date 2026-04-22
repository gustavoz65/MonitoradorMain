//go:build !cgo

package native

import (
	"bytes"
	"hash/crc32"
	"strings"
)

func normalizeASCII(value string) string {
	return strings.ToLower(value)
}

func containsBytes(body, pattern []byte) bool {
	return bytes.Contains(body, pattern)
}

func hashBytes(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}
