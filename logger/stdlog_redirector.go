package logger

import (
	"bytes"
)

type stdlogRedirector struct{}

func (*stdlogRedirector) Write(b []byte) (n int, err error) {
	msg := bytes.Trim(b, " \t\r\n")

	// Extract the log-level indicator; use info by default
	loc := prefixRegex.FindIndex(b)
	level := 'I'
	if loc != nil {
		level = rune(b[loc[0]])
		msg = bytes.Trim(msg[loc[1]:], " \t\r\n")
	}

	// Log with the given level
	switch level {
	case 'D':
		instance.Debug(string(msg))
	case 'I':
		instance.Info(string(msg))
	case 'W':
		instance.Warn(string(msg))
	case 'E':
		instance.Error(string(msg))
	}

	return len(b), nil
}
