package logger

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/rotate"
	"github.com/influxdata/wlog"
)

const (
	LogTargetFile   = "file"
	LogTargetStderr = "stderr"
)

type defaultLogger struct {
	Category string
	Name     string
	Alias    string

	prefix  string
	onError []func()

	logger   *log.Logger
	level    telegraf.LogLevel
	timezone *time.Location
}

// NewLogger creates a new logger instance
func (l *defaultLogger) New(category, name, alias string) telegraf.Logger {
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

	return &defaultLogger{
		Category: category,
		Name:     name,
		Alias:    alias,
		prefix:   prefix,
		level:    l.level,
		logger:   l.logger,
		timezone: l.timezone,
	}
}

func (l *defaultLogger) Close() error {
	writer := l.logger.Writer()

	// Close the writer if possible and avoid closing stderr
	if writer == os.Stderr {
		return nil
	}
	if closer, ok := writer.(io.Closer); ok {
		return closer.Close()
	}

	return errors.New("the underlying writer cannot be closed")
}

// Register a callback triggered when errors are about to be written to the log
func (l *defaultLogger) RegisterErrorCallback(f func()) {
	l.onError = append(l.onError, f)
}

func (l *defaultLogger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

func (l *defaultLogger) Level() telegraf.LogLevel {
	return l.level
}

// Error logging including callbacks
func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Error(args ...interface{}) {
	l.print(telegraf.Error, time.Now(), args...)
	for _, f := range l.onError {
		f()
	}
}

// Warning logging
func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Warn(args ...interface{}) {
	l.print(telegraf.Warn, time.Now(), args...)
}

// Info logging
func (l *defaultLogger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Info(args ...interface{}) {
	l.print(telegraf.Info, time.Now(), args...)
}

// Debug logging, this is suppressed on console
func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Debug(args ...interface{}) {
	l.print(telegraf.Debug, time.Now(), args...)
}

func (l *defaultLogger) print(level telegraf.LogLevel, ts time.Time, args ...interface{}) {
	// Skip all messages with insufficient log-levels
	if level > l.level {
		return
	}
	msg := append([]interface{}{ts.In(l.timezone).Format(time.RFC3339), " ", level.Indicator(), l.prefix}, args...)
	l.logger.Print(msg...)
}

func createDefaultLogger(cfg *Config) (logger, error) {
	// Set the log-level
	switch cfg.logLevel {
	case telegraf.Error:
		wlog.SetLevel(wlog.ERROR)
	case telegraf.Warn:
		wlog.SetLevel(wlog.WARN)
	case telegraf.Info:
		wlog.SetLevel(wlog.INFO)
	case telegraf.Debug:
		wlog.SetLevel(wlog.DEBUG)
	}

	// Setup the writer target
	var writer io.Writer = os.Stderr
	if cfg.LogTarget == "file" && cfg.Logfile != "" {
		w, err := rotate.NewFileWriter(
			cfg.Logfile,
			cfg.RotationInterval,
			cfg.RotationMaxSize,
			cfg.RotationMaxArchives,
		)
		if err != nil {
			return nil, err
		}
		writer = w
	}

	// Get configured timezone
	timezoneName := cfg.LogWithTimezone
	if strings.EqualFold(timezoneName, "local") {
		timezoneName = "Local"
	}
	tz, err := time.LoadLocation(timezoneName)
	if err != nil {
		return nil, errors.New("error while setting logging timezone: " + err.Error())
	}

	// Setup the logger
	l := &defaultLogger{
		level:    cfg.logLevel,
		prefix:   " ",
		logger:   log.New(writer, "", 0),
		timezone: tz,
	}

	return l, nil
}

func init() {
	add("stderr", createDefaultLogger)
	add("file", createDefaultLogger)
}
