package testutil

import (
	"fmt"
	"log"
	"sync"

	"github.com/influxdata/telegraf"
)

var _ telegraf.Logger = &CaptureLogger{}

// CaptureLogger defines a logging structure for plugins.
type CaptureLogger struct {
	Name   string // Name is the plugin name, will be printed in the `[]`.
	errors []string
	sync.Mutex
}

// Errorf logs an error message, patterned after log.Printf.
func (l *CaptureLogger) Errorf(format string, args ...interface{}) {
	s := fmt.Sprintf("E! ["+l.Name+"] "+format, args...)
	l.Lock()
	l.errors = append(l.errors, s)
	l.Unlock()
	log.Print(s)
}

// Error logs an error message, patterned after log.Print.
func (l *CaptureLogger) Error(args ...interface{}) {
	s := fmt.Sprint(append([]interface{}{"E! [" + l.Name + "] "}, args...)...)
	l.Lock()
	l.errors = append(l.errors, s)
	l.Unlock()
	log.Print(s)
}

// Debugf logs a debug message, patterned after log.Printf.
func (l *CaptureLogger) Debugf(format string, args ...interface{}) {
	log.Printf("D! ["+l.Name+"] "+format, args...)
}

// Debug logs a debug message, patterned after log.Print.
func (l *CaptureLogger) Debug(args ...interface{}) {
	log.Print(append([]interface{}{"D! [" + l.Name + "] "}, args...)...)
}

// Warnf logs a warning message, patterned after log.Printf.
func (l *CaptureLogger) Warnf(format string, args ...interface{}) {
	log.Printf("W! ["+l.Name+"] "+format, args...)
}

// Warn logs a warning message, patterned after log.Print.
func (l *CaptureLogger) Warn(args ...interface{}) {
	log.Print(append([]interface{}{"W! [" + l.Name + "] "}, args...)...)
}

// Infof logs an information message, patterned after log.Printf.
func (l *CaptureLogger) Infof(format string, args ...interface{}) {
	log.Printf("I! ["+l.Name+"] "+format, args...)
}

// Info logs an information message, patterned after log.Print.
func (l *CaptureLogger) Info(args ...interface{}) {
	log.Print(append([]interface{}{"I! [" + l.Name + "] "}, args...)...)
}

func (l *CaptureLogger) Errors() []string {
	l.Lock()
	defer l.Unlock()
	e := append([]string{}, l.errors...)
	return e
}

func (l *CaptureLogger) LastError() string {
	l.Lock()
	defer l.Unlock()
	if len(l.errors) > 0 {
		return l.errors[len(l.errors)-1]
	}
	return ""
}
