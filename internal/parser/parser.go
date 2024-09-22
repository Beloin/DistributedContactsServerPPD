package parser

import "strings"

func ReadTillNull(buffer []byte) string {
	var builder strings.Builder
	for i := 0; i < len(buffer); i++ {
		if buffer[i] == '\000' {
			break
		}

		builder.WriteByte(buffer[i])
	}

	return builder.String()
}

func ParseTo32Bits(buffer []byte) uint32 {
	return uint32(buffer[0])<<24 |
		uint32(buffer[1])<<16 |
		uint32(buffer[2])<<8 |
		uint32(buffer[3])
}

func Parse32Bits(n uint32, buffer *[]byte) {
	buf := *buffer
	buf[0] = byte(n >> 24) // Most significant byte
	buf[1] = byte(n >> 16) // Second byte
	buf[2] = byte(n >> 8)  // Third byte
	buf[3] = byte(n)       // Least significant byte
}
