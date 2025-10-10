package testutil

import (
	"log"

	"github.com/influxdata/telegraf"
)

var _ telegraf.Logger = &Logger{}

type Logger struct {
	Name  string // Name is the plugin name, will be printed in the `[]`.
	Quiet bool
}

func (Logger) Level() telegraf.LogLevel {
	// We always want to output at debug level during testing to find issues easier
	return telegraf.Debug
}

// AddAttribute is not supported by the test-logger
func (Logger) AddAttribute(string, interface{}) {}

func (l Logger) Errorf(format string, args ...interface{}) {
	log.Printf("E! ["+l.Name+"] "+format, args...)
}

func (l Logger) Error(args ...interface{}) {
	log.Print(append([]interface{}{"E! [" + l.Name + "] "}, args...)...)
}

func (l Logger) Warnf(format string, args ...interface{}) {
	log.Printf("W! ["+l.Name+"] "+format, args...)
}

func (l Logger) Warn(args ...interface{}) {
	log.Print(append([]interface{}{"W! [" + l.Name + "] "}, args...)...)
}

func (l Logger) Infof(format string, args ...interface{}) {
	if !l.Quiet {
		log.Printf("I! ["+l.Name+"] "+format, args...)
	}
}

func (l Logger) Info(args ...interface{}) {
	if !l.Quiet {
		log.Print(append([]interface{}{"I! [" + l.Name + "] "}, args...)...)
	}
}

func (l Logger) Debugf(format string, args ...interface{}) {
	if !l.Quiet {
		log.Printf("D! ["+l.Name+"] "+format, args...)
	}
}

func (l Logger) Debug(args ...interface{}) {
	if !l.Quiet {
		log.Print(append([]interface{}{"D! [" + l.Name + "] "}, args...)...)
	}
}

func (l Logger) Tracef(format string, args ...interface{}) {
	if !l.Quiet {
		log.Printf("T! ["+l.Name+"] "+format, args...)
	}
}

// Trace logs a trace message, patterned after log.Print.
func (l Logger) Trace(args ...interface{}) {
	if !l.Quiet {
		log.Print(append([]interface{}{"T! [" + l.Name + "] "}, args...)...)
	}
}
