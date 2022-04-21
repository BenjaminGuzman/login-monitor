package main

import (
	"bytes"
	"encoding/base64"
)

func Base64Encode(src []byte) []byte {
	dst := bytes.Buffer{}
	writer := base64.NewEncoder(base64.StdEncoding, &dst)
	_, _ = writer.Write(src)
	_ = writer.Close() // flush the buffer
	return dst.Bytes()
}

func MinInt(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

// Wrap the contents of the byte array writing sep after maxLen bytes. This is repeated until the end is reached
func Wrap(src []byte, maxLen int, sep string) []byte {
	dst := bytes.Buffer{}
	dst.Grow(len(src))
	for i := 0; i < len(src); i += maxLen {
		dst.Write(src[i:MinInt(len(src), i+maxLen)])
		dst.WriteString(sep)
	}
	return dst.Bytes()
}
