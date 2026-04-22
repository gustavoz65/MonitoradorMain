//go:build cgo

package native

/*
#include <ctype.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

static void monimaster_ascii_lower(char *s) {
	while (*s) {
		*s = (char)tolower((unsigned char)*s);
		s++;
	}
}

static int monimaster_contains(const char *hay, size_t hlen, const char *needle, size_t nlen) {
	if (nlen == 0) return 1;
	if (hlen < nlen) return 0;
	for (size_t i = 0; i <= hlen - nlen; i++) {
		if (memcmp(hay + i, needle, nlen) == 0) return 1;
	}
	return 0;
}

static uint32_t monimaster_crc32(const uint8_t *data, size_t len) {
	uint32_t crc = 0xFFFFFFFF;
	for (size_t i = 0; i < len; i++) {
		crc ^= data[i];
		for (int j = 0; j < 8; j++) {
			crc = (crc >> 1) ^ (0xEDB88320u & (uint32_t)(-(int32_t)(crc & 1)));
		}
	}
	return ~crc;
}
*/
import "C"

import "unsafe"

func normalizeASCII(value string) string {
	cstr := C.CString(value)
	defer C.free(unsafe.Pointer(cstr))
	C.monimaster_ascii_lower(cstr)
	return C.GoString(cstr)
}

func containsBytes(body, pattern []byte) bool {
	if len(pattern) == 0 {
		return true
	}
	if len(body) == 0 {
		return false
	}
	return C.monimaster_contains(
		(*C.char)(unsafe.Pointer(&body[0])), C.size_t(len(body)),
		(*C.char)(unsafe.Pointer(&pattern[0])), C.size_t(len(pattern)),
	) == 1
}

func hashBytes(data []byte) uint32 {
	if len(data) == 0 {
		return 0
	}
	return uint32(C.monimaster_crc32((*C.uint8_t)(unsafe.Pointer(&data[0])), C.size_t(len(data))))
}
