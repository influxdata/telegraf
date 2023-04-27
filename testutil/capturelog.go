package testutil

import (
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/influxdata/telegraf"
)

var _ telegraf.Logger = &CaptureLogger{}

const (
	logLevelDebug = 'D'
	logLevelWarn  = 'W'
	logLevelInfo  = 'I'
	logLevelError = 'E'
)

type logMessage struct {
	level   byte
	message string
}

// CaptureLogger defines a logging structure for plugins.
type CaptureLogger struct {
	T        *testing.T
	Name     string // Name is the plugin name, will be printed in the `[]`.
	messages []logMessage
	sync.Mutex
}

func (l *CaptureLogger) msgString(msg logMessage) string {
	return fmt.Sprintf("%c! [%s] %s", msg.level, l.Name, msg.message)
}
func (l *CaptureLogger) logMsg(msg logMessage) {
	l.Lock()
	l.messages = append(l.messages, msg)
	l.T.Log(l.msgString(msg))
	l.Unlock()
}

func (l *CaptureLogger) logf(level byte, format string, args ...any) {
	l.logMsg(logMessage{level, fmt.Sprintf(format, args...)})
}

func (l *CaptureLogger) loga(level byte, args ...any) {
	l.logMsg(logMessage{level, fmt.Sprint(args...)})
}

// Errorf logs an error message, patterned after log.Printf.
func (l *CaptureLogger) Errorf(format string, args ...interface{}) {
	l.logf(logLevelError, format, args...)
}

// Error logs an error message, patterned after log.Print.
func (l *CaptureLogger) Error(args ...interface{}) {
	l.loga(logLevelError, args...)
}

// Debugf logs a debug message, patterned after log.Printf.
func (l *CaptureLogger) Debugf(format string, args ...interface{}) {
	l.logf(logLevelDebug, format, args...)
}

// Debug logs a debug message, patterned after log.Print.
func (l *CaptureLogger) Debug(args ...interface{}) {
	l.loga(logLevelDebug, args...)
}

// Warnf logs a warning message, patterned after log.Printf.
func (l *CaptureLogger) Warnf(format string, args ...interface{}) {
	log.Printf("W! ["+l.Name+"] "+format, args...)
	l.logf(logLevelWarn, format, args...)
}

// Warn logs a warning message, patterned after log.Print.
func (l *CaptureLogger) Warn(args ...interface{}) {
	l.loga(logLevelWarn, args...)
}

// Infof logs an information message, patterned after log.Printf.
func (l *CaptureLogger) Infof(format string, args ...interface{}) {
	l.logf(logLevelInfo, format, args...)
}

// Info logs an information message, patterned after log.Print.
func (l *CaptureLogger) Info(args ...interface{}) {
	l.loga(logLevelInfo, args...)
}

func (l *CaptureLogger) Messages() []string {
	l.Lock()
	msgs := make([]string, len(l.messages))
	for i, m := range l.messages {
		msgs[i] = l.msgString(m)
	}
	l.Unlock()
	return msgs
}

func (l *CaptureLogger) filter(level byte) []string {
	l.Lock()
	defer l.Unlock()
	var msgs []string
	for _, m := range l.messages {
		if m.level == level {
			msgs = append(msgs, m.message)
		}
	}
	return msgs
}

func (l *CaptureLogger) Errors() []string {
	return l.filter(logLevelError)
}

func (l *CaptureLogger) Warns() []string {
	return l.filter(logLevelWarn)
}

func (l *CaptureLogger) LastError() string {
	l.Lock()
	defer l.Unlock()
	for i := len(l.messages) - 1; i >= 0; i-- {
		if l.messages[i].level == logLevelError {
			return l.messages[i].message
		}
	}
	return ""
}
