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
)

const (
	LogTargetFile   = "file"
	LogTargetStderr = "stderr"
)

type defaultLogger struct {
	prefix  string
	onError []func()

	logger   *log.Logger
	level    telegraf.LogLevel
	timezone *time.Location
}

// NewLogger creates a new logger instance
func (l *defaultLogger) New(tag string) telegraf.Logger {
	prefix := l.prefix
	if prefix != "" && tag != "" {
		prefix += "." + tag
	} else {
		prefix = tag
	}
	return &defaultLogger{
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
	l.Print(telegraf.Error, time.Now(), args...)
	for _, f := range l.onError {
		f()
	}
}

// Warning logging
func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Warn(args ...interface{}) {
	l.Print(telegraf.Warn, time.Now(), args...)
}

// Info logging
func (l *defaultLogger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Info(args ...interface{}) {
	l.Print(telegraf.Info, time.Now(), args...)
}

// Debug logging, this is suppressed on console
func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l *defaultLogger) Debug(args ...interface{}) {
	l.Print(telegraf.Debug, time.Now(), args...)
}

func (l *defaultLogger) Print(level telegraf.LogLevel, ts time.Time, args ...interface{}) {
	// Skip all messages with insufficient log-levels
	if level > l.level {
		return
	}
	var prefix string
	if l.prefix != "" {
		prefix = "[" + l.prefix + "] "
	}
	msg := append([]interface{}{ts.In(l.timezone).Format(time.RFC3339), " ", level.Indicator(), " ", prefix}, args...)
	l.logger.Print(msg...)
}

func createDefaultLogger(cfg *Config) (logger, error) {
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
		logger:   log.New(writer, "", 0),
		timezone: tz,
	}

	return l, nil
}

func init() {
	add("stderr", createDefaultLogger)
	add("file", createDefaultLogger)
}
