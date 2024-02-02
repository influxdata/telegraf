package testutil

import (
	"fmt"
	"log"
	"sync"

	"github.com/influxdata/telegraf"
)

var _ telegraf.Logger = &CaptureLogger{}

const (
	LevelError = 'E'
	LevelWarn  = 'W'
	LevelInfo  = 'I'
	LevelDebug = 'D'
)

type Entry struct {
	Level byte
	Name  string
	Text  string
}

func (e *Entry) String() string {
	return fmt.Sprintf("%c! [%s] %s", e.Level, e.Name, e.Text)
}

// CaptureLogger defines a logging structure for plugins.
type CaptureLogger struct {
	Name     string // Name is the plugin name, will be printed in the `[]`.
	messages []Entry
	sync.Mutex
}

func (l *CaptureLogger) print(msg Entry) {
	l.Lock()
	l.messages = append(l.messages, msg)
	l.Unlock()
	log.Print(msg.String())
}

func (l *CaptureLogger) logf(level byte, format string, args ...any) {
	l.print(Entry{level, l.Name, fmt.Sprintf(format, args...)})
}

func (l *CaptureLogger) loga(level byte, args ...any) {
	l.print(Entry{level, l.Name, fmt.Sprint(args...)})
}

// Errorf logs an error message, patterned after log.Printf.
func (l *CaptureLogger) Errorf(format string, args ...interface{}) {
	l.logf(LevelError, format, args...)
}

// Error logs an error message, patterned after log.Print.
func (l *CaptureLogger) Error(args ...interface{}) {
	l.loga(LevelError, args...)
}

// Debugf logs a debug message, patterned after log.Printf.
func (l *CaptureLogger) Debugf(format string, args ...interface{}) {
	l.logf(LevelDebug, format, args...)
}

// Debug logs a debug message, patterned after log.Print.
func (l *CaptureLogger) Debug(args ...interface{}) {
	l.loga(LevelDebug, args...)
}

// Warnf logs a warning message, patterned after log.Printf.
func (l *CaptureLogger) Warnf(format string, args ...interface{}) {
	l.logf(LevelWarn, format, args...)
}

// Warn logs a warning message, patterned after log.Print.
func (l *CaptureLogger) Warn(args ...interface{}) {
	l.loga(LevelWarn, args...)
}

// Infof logs an information message, patterned after log.Printf.
func (l *CaptureLogger) Infof(format string, args ...interface{}) {
	l.logf(LevelInfo, format, args...)
}

// Info logs an information message, patterned after log.Print.
func (l *CaptureLogger) Info(args ...interface{}) {
	l.loga(LevelInfo, args...)
}

func (l *CaptureLogger) Messages() []Entry {
	l.Lock()
	msgs := make([]Entry, len(l.messages))
	copy(msgs, l.messages)
	l.Unlock()
	return msgs
}

func (l *CaptureLogger) filter(level byte) []string {
	l.Lock()
	defer l.Unlock()
	var msgs []string
	for _, m := range l.messages {
		if m.Level == level {
			msgs = append(msgs, m.String())
		}
	}
	return msgs
}

func (l *CaptureLogger) Errors() []string {
	return l.filter(LevelError)
}

func (l *CaptureLogger) Warnings() []string {
	return l.filter(LevelWarn)
}

func (l *CaptureLogger) LastError() string {
	l.Lock()
	defer l.Unlock()
	for i := len(l.messages) - 1; i >= 0; i-- {
		if l.messages[i].Level == LevelError {
			return l.messages[i].String()
		}
	}
	return ""
}

func (l *CaptureLogger) Clear() {
	l.Lock()
	defer l.Unlock()
	l.messages = make([]Entry, 0)
}
