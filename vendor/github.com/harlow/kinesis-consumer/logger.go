package consumer

import (
	"log"
)

// A Logger is a minimal interface to as a adaptor for external logging library to consumer
type Logger interface {
	Log(...interface{})
}

// noopLogger implements logger interface with discard
type noopLogger struct {
	logger *log.Logger
}

// Log using stdlib logger. See log.Println.
func (l noopLogger) Log(args ...interface{}) {
	l.logger.Println(args...)
}
