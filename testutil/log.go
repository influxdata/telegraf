package testutil

import (
	"log"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
)

var _ telegraf.Logger = &Logger{}

// Logger defines a logging structure for plugins.
type Logger struct {
	Name string // Name is the plugin name, will be printed in the `[]`.
}

// Errorf logs an error message, patterned after log.Printf.
func (l Logger) Errorf(format string, args ...interface{}) {
	log.Printf("E! ["+l.Name+"] "+format, internal.SanitizeArgs(args)...)
}

// Error logs an error message, patterned after log.Print.
func (l Logger) Error(args ...interface{}) {
	log.Print(append([]interface{}{"E! [" + l.Name + "] "}, args...)...)
}

// Debugf logs a debug message, patterned after log.Printf.
func (l Logger) Debugf(format string, args ...interface{}) {
	log.Printf("D! ["+l.Name+"] "+format, internal.SanitizeArgs(args)...)
}

// Debug logs a debug message, patterned after log.Print.
func (l Logger) Debug(args ...interface{}) {
	log.Print(append([]interface{}{"D! [" + l.Name + "] "}, args...)...)
}

// Warnf logs a warning message, patterned after log.Printf.
func (l Logger) Warnf(format string, args ...interface{}) {
	log.Printf("W! ["+l.Name+"] "+format, internal.SanitizeArgs(args)...)
}

// Warn logs a warning message, patterned after log.Print.
func (l Logger) Warn(args ...interface{}) {
	log.Print(append([]interface{}{"W! [" + l.Name + "] "}, args...)...)
}

// Infof logs an information message, patterned after log.Printf.
func (l Logger) Infof(format string, args ...interface{}) {
	log.Printf("I! ["+l.Name+"] "+format, internal.SanitizeArgs(args)...)
}

// Info logs an information message, patterned after log.Print.
func (l Logger) Info(args ...interface{}) {
	log.Print(append([]interface{}{"I! [" + l.Name + "] "}, args...)...)
}
