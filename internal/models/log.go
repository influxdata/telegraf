package models

import (
	"log"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
)

// PluginConfig contains individualized plugin configuration.
type PluginConfig struct {
	Log Logger
}

// Logger returns PluginConfig's Logger in order to satisfy the interface.
func (p PluginConfig) Logger() telegraf.Logger {
	return p.Log
}

// Logger defines a logging structure for plugins.
type Logger struct {
	Name string // Name is the plugin name, will be printed in the `[]`.
}

// Errorf logs an error message, patterned after log.Printf.
func (l Logger) Errorf(format string, args ...interface{}) {
	// todo: keep tally of errors from plugins
	log.Printf("E! ["+l.Name+"] "+format, args...)
}

// Error logs an error message, patterned after log.Print.
func (l Logger) Error(args ...interface{}) {
	log.Print(append([]interface{}{"E! [" + l.Name + "] "}, args...)...)
}

// Debugf logs a debug message, patterned after log.Printf.
func (l Logger) Debugf(format string, args ...interface{}) {
	log.Printf("D! ["+l.Name+"] "+format, args...)
}

// Debug logs a debug message, patterned after log.Print.
func (l Logger) Debug(args ...interface{}) {
	log.Print(append([]interface{}{"D! [" + l.Name + "] "}, args...)...)
}

// Warnf logs a warning message, patterned after log.Printf.
func (l Logger) Warnf(format string, args ...interface{}) {
	log.Printf("W! ["+l.Name+"] "+format, args...)
}

// Warn logs a warning message, patterned after log.Print.
func (l Logger) Warn(args ...interface{}) {
	log.Print(append([]interface{}{"W! [" + l.Name + "] "}, args...)...)
}

// Infof logs an information message, patterned after log.Printf.
func (l Logger) Infof(format string, args ...interface{}) {
	log.Printf("I! ["+l.Name+"] "+format, args...)
}

// Info logs an information message, patterned after log.Print.
func (l Logger) Info(args ...interface{}) {
	log.Print(append([]interface{}{"I! [" + l.Name + "] "}, args...)...)
}

func (l Logger) addError() {
	switch {
	case strings.HasPrefix(l.Name, "aggregator"):
		fallthrough
	case strings.HasPrefix(l.Name, "input"):
		iErrors.Incr(1)
	case strings.HasPrefix(l.Name, "output"):
		oErrors.Incr(1)
	case strings.HasPrefix(l.Name, "processor"):
		pErrors.Incr(1)
	}
}

var (
	aErrors = selfstat.Register("agent", "aggregator_errors", map[string]string{})
	iErrors = selfstat.Register("agent", "input_errors", map[string]string{})
	oErrors = selfstat.Register("agent", "output_errors", map[string]string{})
	pErrors = selfstat.Register("agent", "processor_errors", map[string]string{})
)
