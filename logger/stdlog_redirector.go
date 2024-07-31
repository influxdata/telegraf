package logger

import (
	"bytes"
	"regexp"
)

var prefixRegex = regexp.MustCompile("^[DIWE]!")

type stdlogRedirector struct {
	log logger
}

func (s *stdlogRedirector) Write(b []byte) (n int, err error) {
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
	case 'T':
		s.log.Trace(string(msg))
	case 'D':
		s.log.Debug(string(msg))
	case 'I':
		s.log.Info(string(msg))
	case 'W':
		s.log.Warn(string(msg))
	case 'E':
		s.log.Error(string(msg))
	}

	return len(b), nil
}
