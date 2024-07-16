//go:build windows

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/influxdata/telegraf"
	"golang.org/x/sys/windows/svc/eventlog"
)

const (
	eidInfo    = 1
	eidWarning = 2
	eidError   = 3
)

type eventLogger struct {
	prefix  string
	onError []func()

	eventlog *eventlog.Log
	level    telegraf.LogLevel
	errlog   *log.Logger
}

// NewLogger creates a new logger instance
func (l *eventLogger) New(tag string) telegraf.Logger {
	prefix := l.prefix
	if prefix != "" && tag != "" {
		prefix += "." + tag
	} else {
		prefix = tag
	}

	return &eventLogger{
		prefix:   prefix,
		eventlog: l.eventlog,
		level:    l.level,
		errlog:   l.errlog,
	}
}

func (l *eventLogger) Close() error {
	return l.eventlog.Close()
}

// Register a callback triggered when errors are about to be written to the log
func (l *eventLogger) RegisterErrorCallback(f func()) {
	l.onError = append(l.onError, f)
}

// Redirecting output not supported by eventlog
func (*eventLogger) SetOutput(io.Writer) {}

func (l *eventLogger) Level() telegraf.LogLevel {
	return l.level
}

// Error logging including callbacks
func (l *eventLogger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *eventLogger) Error(args ...interface{}) {
	l.Print(telegraf.Error, time.Now(), args...)
	for _, f := range l.onError {
		f()
	}
}

// Warning logging
func (l *eventLogger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *eventLogger) Warn(args ...interface{}) {
	l.Print(telegraf.Warn, time.Now(), args...)
}

// Info logging
func (l *eventLogger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *eventLogger) Info(args ...interface{}) {
	l.Print(telegraf.Info, time.Now(), args...)
}

// Debug logging is not supported by eventlog
func (*eventLogger) Debugf(string, ...interface{}) {}
func (*eventLogger) Debug(...interface{})          {}

func (l *eventLogger) Print(level telegraf.LogLevel, _ time.Time, args ...interface{}) {
	// Skip all messages with insufficient log-levels
	if level > l.level {
		return
	}

	var prefix string
	if l.prefix != "" {
		prefix = "[" + l.prefix + "] "
	}
	msg := level.Indicator() + " " + prefix + fmt.Sprint(args...)

	var err error
	switch level {
	case telegraf.Error:
		err = l.eventlog.Error(eidError, msg)
	case telegraf.Warn:
		err = l.eventlog.Warning(eidWarning, msg)
	case telegraf.Info:
		err = l.eventlog.Info(eidInfo, msg)
	}
	if err != nil {
		log.Printf("E! Writing log message failed: %v", err)
	}
}

func createEventLogger(cfg *Config) (logger, error) {
	eventLog, err := eventlog.Open(cfg.InstanceName)
	if err != nil {
		return nil, err
	}

	l := &eventLogger{
		eventlog: eventLog,
		level:    cfg.logLevel,
		errlog:   log.New(os.Stderr, "", 0),
	}

	return l, nil
}

func init() {
	add("eventlog", createEventLogger)
}
