package logger

import (
	"fmt"
	"log"
)

// Logger is common logging interface.
type Logger interface {
	Errorf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

// Level is logging severity.
type Level uint8

const (
	LevelOff Level = iota
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
)

// String returns log level string representation.
func (lvl Level) String() string {
	switch lvl {
	case LevelOff:
		return "OFF"
	case LevelError:
		return "ERROR"
	case LevelWarn:
		return "WARN"
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", lvl)
	}
}

// OutputFunc prints logs.
type OutputFunc func(lvl Level, s string)

// New creates a new leveled logger instance with the given parameters.
//
// If out is nil it uses the standard logger for output.
func New(lvl Level, out OutputFunc) *LevelLogger {
	if out == nil {
		out = logStd
	}
	return &LevelLogger{lvl: lvl, out: out}
}

func logStd(lvl Level, s string) {
	_ = log.Output(4, fmt.Sprint(lvl.String(), " ", s))
}

// LevelLogger is a logger that supports log levels.
type LevelLogger struct {
	lvl Level
	out OutputFunc
}

func (l *LevelLogger) Errorf(format string, v ...interface{}) {
	l.logf(LevelError, format, v...)
}

func (l *LevelLogger) Infof(format string, v ...interface{}) {
	l.logf(LevelInfo, format, v...)
}

func (l *LevelLogger) Warnf(format string, v ...interface{}) {
	l.logf(LevelWarn, format, v...)
}

func (l *LevelLogger) Debugf(format string, v ...interface{}) {
	l.logf(LevelDebug, format, v...)
}

func (l *LevelLogger) logf(lvl Level, format string, v ...interface{}) {
	if lvl <= l.lvl {
		l.out(lvl, fmt.Sprintf(format, v...))
	}
}
