package models

import (
	"log"
	"reflect"

	"github.com/influxdata/telegraf"
)

// Logger defines a logging structure for plugins.
type Logger struct {
	OnErrs []func()
	Name   string // Name is the plugin name, will be printed in the `[]`.
}

// NewLogger creates a new logger instance
func NewLogger(pluginType, name, alias string) *Logger {
	return &Logger{
		Name: logName(pluginType, name, alias),
	}
}

// OnErr defines a callback that triggers only when errors are about to be written to the log
func (l *Logger) OnErr(f func()) {
	l.OnErrs = append(l.OnErrs, f)
}

// Errorf logs an error message, patterned after log.Printf.
func (l *Logger) Errorf(format string, args ...interface{}) {
	for _, f := range l.OnErrs {
		f()
	}
	log.Printf("E! ["+l.Name+"] "+format, args...)
}

// Error logs an error message, patterned after log.Print.
func (l *Logger) Error(args ...interface{}) {
	for _, f := range l.OnErrs {
		f()
	}
	log.Print(append([]interface{}{"E! [" + l.Name + "] "}, args...)...)
}

// Debugf logs a debug message, patterned after log.Printf.
func (l *Logger) Debugf(format string, args ...interface{}) {
	log.Printf("D! ["+l.Name+"] "+format, args...)
}

// Debug logs a debug message, patterned after log.Print.
func (l *Logger) Debug(args ...interface{}) {
	log.Print(append([]interface{}{"D! [" + l.Name + "] "}, args...)...)
}

// Warnf logs a warning message, patterned after log.Printf.
func (l *Logger) Warnf(format string, args ...interface{}) {
	log.Printf("W! ["+l.Name+"] "+format, args...)
}

// Warn logs a warning message, patterned after log.Print.
func (l *Logger) Warn(args ...interface{}) {
	log.Print(append([]interface{}{"W! [" + l.Name + "] "}, args...)...)
}

// Infof logs an information message, patterned after log.Printf.
func (l *Logger) Infof(format string, args ...interface{}) {
	log.Printf("I! ["+l.Name+"] "+format, args...)
}

// Info logs an information message, patterned after log.Print.
func (l *Logger) Info(args ...interface{}) {
	log.Print(append([]interface{}{"I! [" + l.Name + "] "}, args...)...)
}

// logName returns the log-friendly name/type.
func logName(pluginType, name, alias string) string {
	if alias == "" {
		return pluginType + "." + name
	}
	return pluginType + "." + name + "::" + alias
}

func SetLoggerOnPlugin(i interface{}, logger telegraf.Logger) {
	valI := reflect.ValueOf(i)

	if valI.Type().Kind() != reflect.Ptr {
		valI = reflect.New(reflect.TypeOf(i))
	}

	field := valI.Elem().FieldByName("Log")
	if !field.IsValid() {
		return
	}

	switch field.Type().String() {
	case "telegraf.Logger":
		if field.CanSet() {
			field.Set(reflect.ValueOf(logger))
		}
	default:
		logger.Debugf("Plugin %q defines a 'Log' field on its struct of an unexpected type %q. Expected telegraf.Logger",
			valI.Type().Name(), field.Type().String())
	}
}
