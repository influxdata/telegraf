//go:build windows

package logger

import (
	"fmt"
	"log"
	"strings"

	"github.com/influxdata/telegraf"
	"golang.org/x/sys/windows/svc/eventlog"
)

const (
	eidInfo    = 1
	eidWarning = 2
	eidError   = 3
)

type eventLogger struct {
	Category string
	Name     string
	Alias    string
	LogLevel telegraf.LogLevel

	prefix  string
	onError []func()

	eventlog *eventlog.Log
}

func (e *eventLogger) Write(b []byte) (int, error) {
	loc := prefixRegex.FindIndex(b)
	n := len(b)
	if loc == nil {
		return n, e.eventlog.Info(1, string(b))
	}

	// Skip empty log messages
	if n <= 2 {
		return 0, nil
	}

	line := strings.Trim(string(b[loc[1]:]), " \t\r\n")
	switch rune(b[loc[0]]) {
	case 'I':
		return n, e.eventlog.Info(eidInfo, line)
	case 'W':
		return n, e.eventlog.Warning(eidWarning, line)
	case 'E':
		return n, e.eventlog.Error(eidError, line)
	}

	return n, nil
}

// NewLogger creates a new logger instance
func (e *eventLogger) New(category, name, alias string) telegraf.Logger {
	var prefix string
	if category != "" {
		prefix = "[" + category
		if name != "" {
			prefix += "." + name
		}
		if alias != "" {
			prefix += "::" + alias
		}
		prefix += "] "
	}

	return &eventLogger{
		Category: category,
		Name:     name,
		Alias:    alias,
		LogLevel: e.LogLevel,
		prefix:   prefix,
		eventlog: e.eventlog,
	}
}

func (e *eventLogger) Close() error {
	return e.eventlog.Close()
}

// OnErr defines a callback that triggers only when errors are about to be written to the log
func (e *eventLogger) RegisterErrorCallback(f func()) {
	e.onError = append(e.onError, f)
}

func (e *eventLogger) Level() telegraf.LogLevel {
	return e.LogLevel
}

// Errorf logs an error message, patterned after log.Printf.
func (e *eventLogger) Errorf(format string, args ...interface{}) {
	e.Error(fmt.Sprintf(format, args...))
}

// Error logs an error message, patterned after log.Print.
func (e *eventLogger) Error(args ...interface{}) {
	if e.LogLevel >= telegraf.Error {
		if err := e.eventlog.Error(eidError, "E! "+e.prefix+fmt.Sprint(args...)); err != nil {
			log.Printf("E! Writing log message failed: %v", err)
		}
	}

	for _, f := range e.onError {
		f()
	}
}

// Warnf logs a warning message, patterned after log.Printf.
func (e *eventLogger) Warnf(format string, args ...interface{}) {
	e.Warn(fmt.Sprintf(format, args...))
}

// Warn logs a warning message, patterned after log.Print.
func (e *eventLogger) Warn(args ...interface{}) {
	if e.LogLevel < telegraf.Warn {
		return
	}
	if err := e.eventlog.Warning(eidError, "W! "+e.prefix+fmt.Sprint(args...)); err != nil {
		log.Printf("E! Writing log message failed: %v", err)
	}
}

// Infof logs an information message, patterned after log.Printf.
func (e *eventLogger) Infof(format string, args ...interface{}) {
	e.Info(fmt.Sprintf(format, args...))
}

// Info logs an information message, patterned after log.Print.
func (e *eventLogger) Info(args ...interface{}) {
	if e.LogLevel < telegraf.Info {
		return
	}
	if err := e.eventlog.Info(eidError, "I! "+e.prefix+fmt.Sprint(args...)); err != nil {
		log.Printf("E! Writing log message failed: %v", err)
	}
}

// No debugging output for eventlog to not spam the service
func (e *eventLogger) Debugf(string, ...interface{}) {}

// No debugging output for eventlog to not spam the service
func (e *eventLogger) Debug(...interface{}) {}

func createEventLogger(cfg *Config) (logger, error) {
	eventLog, err := eventlog.Open(cfg.InstanceName)
	if err != nil {
		return nil, err
	}

	l := &eventLogger{
		eventlog: eventLog,
	}

	log.SetOutput(l)

	return l, nil
}

func init() {
	add("eventlog", createEventLogger)
}
