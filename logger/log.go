package logger

import (
	"log"

	"github.com/influxdata/telegraf"
)

// Logger defines a logging structure for plugins.
type Logger struct {
	Category string
	Name     string
	Alias    string
	LogLevel telegraf.LogLevel

	prefix  string
	onError []func()
}

// NewLogger creates a new logger instance
func NewLogger(category, name, alias string) telegraf.Logger {
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

	return &Logger{
		Category: category,
		Name:     name,
		Alias:    alias,
		LogLevel: telegraf.Info,
		prefix:   prefix,
	}
}

// OnErr defines a callback that triggers only when errors are about to be written to the log
func (l *Logger) RegisterErrorCallback(f func()) {
	l.onError = append(l.onError, f)
}

func (l *Logger) Level() telegraf.LogLevel {
	return l.LogLevel
}

// Errorf logs an error message, patterned after log.Printf.
func (l *Logger) Errorf(format string, args ...interface{}) {
	log.Printf("E! "+l.prefix+format, args...)
	for _, f := range l.onError {
		f()
	}
}

// Error logs an error message, patterned after log.Print.
func (l *Logger) Error(args ...interface{}) {
	for _, f := range l.onError {
		f()
	}
	log.Print(append([]interface{}{"E! " + l.prefix}, args...)...)
}

// Debugf logs a debug message, patterned after log.Printf.
func (l *Logger) Debugf(format string, args ...interface{}) {
	log.Printf("D! "+l.prefix+" "+format, args...)
}

// Debug logs a debug message, patterned after log.Print.
func (l *Logger) Debug(args ...interface{}) {
	log.Print(append([]interface{}{"D! " + l.prefix}, args...)...)
}

// Warnf logs a warning message, patterned after log.Printf.
func (l *Logger) Warnf(format string, args ...interface{}) {
	log.Printf("W! "+l.prefix+format, args...)
}

// Warn logs a warning message, patterned after log.Print.
func (l *Logger) Warn(args ...interface{}) {
	log.Print(append([]interface{}{"W! " + l.prefix}, args...)...)
}

// Infof logs an information message, patterned after log.Printf.
func (l *Logger) Infof(format string, args ...interface{}) {
	log.Printf("I! "+l.prefix+format, args...)
}

// Info logs an information message, patterned after log.Print.
func (l *Logger) Info(args ...interface{}) {
	log.Print(append([]interface{}{"I! " + l.prefix}, args...)...)
}
