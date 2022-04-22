package main

import (
	"bytes"
	"encoding/base64"
	"os"
	"strings"
	"time"
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

func replaceTimePlaceHolder(str, sepStart, sepEnd string) string {
	var strBuilder strings.Builder
	var tokenStart, stringStart int = 0, 0
	var ignoreStart, ignoreEnd int = 0, 0

	for stringStart < len(str) {
		ignoreStart = stringStart + strings.Index(str[stringStart:], sepStart) // str[ignoreStart:] = "%t<format>t%..." (assume sepStart = "%t", sepEnd="t%")
		if ignoreStart == stringStart+-1 {
			break
		}
		tokenStart = ignoreStart + len(sepStart) // str[tokenStart:] = "<format>t%..."

		tokenEnd := tokenStart + strings.Index(str[tokenStart:], sepEnd) // str[:tokenEnd] = "...%t<format>"
		if tokenEnd == tokenStart+-1 {
			break
		}
		ignoreEnd = tokenEnd + len(sepEnd) // str[:ignoreEnd] = "...%t<format>t%"

		strBuilder.WriteString(str[stringStart:ignoreStart])

		format := strings.TrimSpace(str[tokenStart:tokenEnd])
		switch format { // replace special values
		case "ANSIC":
			format = time.ANSIC
		case "UnixDate":
			format = time.UnixDate
		case "RubyDate":
			format = time.RubyDate
		case "RFC822":
			format = time.RFC822
		case "RFC822Z":
			format = time.RFC822Z
		}

		strBuilder.WriteString(time.Now().Format(format))
		stringStart = ignoreEnd // ignore the replaced text in the next loop (str is never modified)
	}
	strBuilder.WriteString(str[stringStart:])

	return strBuilder.String()
}

func replaceFilePlaceHolder(str, sepStart, sepEnd string) string {
	var strBuilder strings.Builder
	var tokenStart, stringStart int = 0, 0
	var ignoreStart, ignoreEnd int = 0, 0

	for stringStart < len(str) {
		ignoreStart = stringStart + strings.Index(str[stringStart:], sepStart) // str[ignoreStart:] = "%f<file>f%..." (assume sepStart = "%f", sepEnd="f%")
		if ignoreStart == stringStart+-1 {
			break
		}
		tokenStart = ignoreStart + len(sepStart) // str[tokenStart:] = "<file>f%..."

		tokenEnd := tokenStart + strings.Index(str[tokenStart:], sepEnd) // str[:tokenEnd] = "...%f<file>"
		if tokenEnd == tokenStart+-1 {
			break
		}
		ignoreEnd = tokenEnd + len(sepEnd) // str[:ignoreEnd] = "...%f<file>f%"

		strBuilder.WriteString(str[stringStart:ignoreStart])

		filePath := strings.TrimSpace(str[tokenStart:tokenEnd])
		if fileContents, err := os.ReadFile(filePath); err == nil {
			strBuilder.Write(fileContents)
		}
		stringStart = ignoreEnd // ignore the replaced text in the next loop (str is never modified)
	}
	strBuilder.WriteString(str[stringStart:])

	return strBuilder.String()
}

// ReplacePlaceholders Replaces
// %h for the hostname
// %t<time format>t% for time.Now().Format(<time format>)
// %f<file path>f% for the contents of <file path> (read permission is required)
func ReplacePlaceholders(str string) string {
	// replace hostname
	if hostname, err := os.Hostname(); err == nil {
		str = strings.ReplaceAll(str, "%h", hostname)
	}

	// replace time
	const timeSepStart = "%t"
	const timeSepEnd = "t%"
	str = replaceTimePlaceHolder(str, timeSepStart, timeSepEnd)

	// replace file
	const fileSepStart = "%f"
	const fileSepEnd = "f%"
	str = replaceFilePlaceHolder(str, fileSepStart, fileSepEnd)

	return str
}
