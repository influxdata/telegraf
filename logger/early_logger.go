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

type msgbuffer struct {
	entries *list.List
	sync.Mutex
}

func (b *msgbuffer) add(level telegraf.LogLevel, ts time.Time, prefix string, args ...interface{}) *entry {
	e := &entry{
		timestamp: ts,
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

	logger *log.Logger
	level  telegraf.LogLevel
	buffer *msgbuffer
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
		level:  l.level,
		prefix: prefix,
		logger: l.logger,
		buffer: l.buffer,
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
	return l.level
}

// Error logging including callbacks
func (l *earlyLogger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *earlyLogger) Error(args ...interface{}) {
	l.Print(telegraf.Error, time.Now(), args...)
	for _, f := range l.onError {
		f()
	}
}

// Warning logging
func (l *earlyLogger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *earlyLogger) Warn(args ...interface{}) {
	l.Print(telegraf.Warn, time.Now(), args...)
}

// Info logging
func (l *earlyLogger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *earlyLogger) Info(args ...interface{}) {
	l.Print(telegraf.Info, time.Now(), args...)
}

// Debug logging
func (l *earlyLogger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l *earlyLogger) Debug(args ...interface{}) {
	l.Print(telegraf.Debug, time.Now(), args...)
}

func (l *earlyLogger) Print(level telegraf.LogLevel, ts time.Time, args ...interface{}) {
	l.buffer.add(level, ts, l.prefix, args...)

	if level <= l.level {
		msg := append([]interface{}{ts.UTC().Format(time.RFC3339), " ", level.Indicator(), " ", l.prefix}, args...)
		l.logger.Print(msg...)
	}
}

// Create an early logging instance
func createEarlyLogger() logger {
	// Setup the logger
	return &earlyLogger{
		level:  telegraf.Info,
		buffer: &msgbuffer{entries: list.New()},
		logger: log.New(os.Stderr, "", 0),
	}
}
