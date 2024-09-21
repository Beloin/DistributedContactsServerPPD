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

