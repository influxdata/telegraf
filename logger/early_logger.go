package logger

import (
	"container/list"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
)

type entry struct {
	timestamp time.Time
	level     telegraf.LogLevel
	args      []interface{}
}

func (e *entry) print(l *log.Logger) {
	msg := append([]interface{}{e.timestamp.UTC().Format(time.RFC3339), " ", e.level.Indicator(), " "}, e.args...)
	l.Print(msg...)
}

type msgbuffer struct {
	entries *list.List
	sync.Mutex
}

func (b *msgbuffer) add(level telegraf.LogLevel, prefix string, args ...interface{}) *entry {
	e := &entry{
		timestamp: time.Now(),
		level:     level,
		args:      append([]interface{}{prefix}, args...),
	}
	b.Lock()
	b.entries.PushBack(e)
	b.Unlock()

	return e
}

type earlyLogger struct {
	prefix  string
	onError []func()

	logger  *log.Logger
	entries *msgbuffer
}

func (l *earlyLogger) New(category, name, alias string) telegraf.Logger {
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
	return &earlyLogger{
		prefix:  prefix,
		logger:  l.logger,
		entries: l.entries,
	}
}

func (*earlyLogger) Close() error {
	return nil
}

func (l *earlyLogger) RegisterErrorCallback(f func()) {
	l.onError = append(l.onError, f)
}

func (l *earlyLogger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

func (l *earlyLogger) Level() telegraf.LogLevel {
	return telegraf.Info
}

// Error logging including callbacks
func (l *earlyLogger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *earlyLogger) Error(args ...interface{}) {
	l.entries.add(telegraf.Error, l.prefix, args...).print(l.logger)
	for _, f := range l.onError {
		f()
	}
}

// Warning logging
func (l *earlyLogger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *earlyLogger) Warn(args ...interface{}) {
	l.entries.add(telegraf.Warn, l.prefix, args...).print(l.logger)
}

// Info logging
func (l *earlyLogger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *earlyLogger) Info(args ...interface{}) {
	l.entries.add(telegraf.Info, l.prefix, args...).print(l.logger)
}

// Debug logging, this is suppressed on console
func (l *earlyLogger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l *earlyLogger) Debug(args ...interface{}) {
	l.entries.add(telegraf.Debug, l.prefix, args...)
}

// Create an early logging instance
func createEarlyLogger() logger {
	// Setup the logger
	return &earlyLogger{
		entries: &msgbuffer{entries: list.New()},
		logger:  log.New(os.Stderr, "", 0),
	}
}
