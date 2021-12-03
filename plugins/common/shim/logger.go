package shim

import (
	"fmt"
	"log" //nolint:revive // Allow exceptional but valid use of log here.
	"os"
	"reflect"

	"github.com/influxdata/telegraf"
)

func init() {
	log.SetOutput(os.Stderr)
}

// Logger defines a logging structure for plugins.
// external plugins can only ever write to stderr and writing to stdout
// would interfere with input/processor writing out of metrics.
type Logger struct{}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{}
}

// Errorf logs an error message, patterned after log.Printf.
func (l *Logger) Errorf(format string, args ...interface{}) {
	log.Printf("E! "+format, args...)
}

// Error logs an error message, patterned after log.Print.
func (l *Logger) Error(args ...interface{}) {
	log.Print("E! ", fmt.Sprint(args...))
}

// Debugf logs a debug message, patterned after log.Printf.
func (l *Logger) Debugf(format string, args ...interface{}) {
	log.Printf("D! "+format, args...)
}

// Debug logs a debug message, patterned after log.Print.
func (l *Logger) Debug(args ...interface{}) {
	log.Print("D! ", fmt.Sprint(args...))
}

// Warnf logs a warning message, patterned after log.Printf.
func (l *Logger) Warnf(format string, args ...interface{}) {
	log.Printf("W! "+format, args...)
}

// Warn logs a warning message, patterned after log.Print.
func (l *Logger) Warn(args ...interface{}) {
	log.Print("W! ", fmt.Sprint(args...))
}

// Infof logs an information message, patterned after log.Printf.
func (l *Logger) Infof(format string, args ...interface{}) {
	log.Printf("I! "+format, args...)
}

// Info logs an information message, patterned after log.Print.
func (l *Logger) Info(args ...interface{}) {
	log.Print("I! ", fmt.Sprint(args...))
}

// setLoggerOnPlugin injects the logger into the plugin,
// if it defines Log telegraf.Logger. This is sort of like SetLogger but using
// reflection instead of forcing the plugin author to define the function for it
func setLoggerOnPlugin(i interface{}, logger telegraf.Logger) {
	valI := reflect.ValueOf(i)

	if valI.Type().Kind() != reflect.Ptr {
		valI = reflect.New(reflect.TypeOf(i))
	}

	field := valI.Elem().FieldByName("Log")
	if !field.IsValid() {
		return
	}

	if field.Type().String() == "telegraf.Logger" {
		if field.CanSet() {
			field.Set(reflect.ValueOf(logger))
		}
	}
}
