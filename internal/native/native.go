package native

func NormalizeASCII(value string) string {
	return normalizeASCII(value)
}

func ContainsBytes(body, pattern []byte) bool {
	return containsBytes(body, pattern)
}

func HashBytes(data []byte) uint32 {
	return hashBytes(data)
}
