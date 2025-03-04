package logger

import (
	"bytes"
)

type stdlogRedirector struct {
	log logger
}

func (s *stdlogRedirector) Write(b []byte) (n int, err error) {
	msg := bytes.Trim(b, " \t\r\n")

	// Check a potential log-level indicator  and log with the given level or
	// use info by default
	switch {
	case bytes.HasPrefix(msg, []byte("E! ")):
		s.log.Error(string(msg[3:]))
	case bytes.HasPrefix(msg, []byte("W! ")):
		s.log.Warn(string(msg[3:]))
	case bytes.HasPrefix(msg, []byte("I! ")):
		s.log.Info(string(msg[3:]))
	case bytes.HasPrefix(msg, []byte("D! ")):
		s.log.Debug(string(msg[3:]))
	case bytes.HasPrefix(msg, []byte("T! ")):
		s.log.Trace(string(msg[3:]))
	default:
		s.log.Info(string(msg))
	}

	return len(b), nil
}
