package logger

import (
	"errors"
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
	LogLevel telegraf.LogLevel

	prefix  string
	onError []func()

	writer         io.Writer
	internalWriter io.Writer
	timezone       *time.Location
}

func (t *defaultLogger) Write(b []byte) (n int, err error) {
	var line []byte
	timeToPrint := time.Now().In(t.timezone)

	if !prefixRegex.Match(b) {
		line = append([]byte(timeToPrint.Format(time.RFC3339)+" I! "), b...)
	} else {
		line = append([]byte(timeToPrint.Format(time.RFC3339)+" "), b...)
	}

	return t.writer.Write(line)
}

// NewLogger creates a new logger instance
func (t *defaultLogger) New(category, name, alias string) telegraf.Logger {
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
		Category:       category,
		Name:           name,
		Alias:          alias,
		LogLevel:       t.LogLevel,
		prefix:         prefix,
		writer:         t.writer,
		internalWriter: t.internalWriter,
		timezone:       t.timezone,
	}
}

func (t *defaultLogger) Close() error {
	// avoid closing stderr
	if t.internalWriter == os.Stderr {
		return nil
	}

	closer, isCloser := t.internalWriter.(io.Closer)
	if !isCloser {
		return errors.New("the underlying writer cannot be closed")
	}
	return closer.Close()
}

// OnErr defines a callback that triggers only when errors are about to be written to the log
func (t *defaultLogger) RegisterErrorCallback(f func()) {
	t.onError = append(t.onError, f)
}

func (t *defaultLogger) Level() telegraf.LogLevel {
	return t.LogLevel
}

// Errorf logs an error message, patterned after log.Printf.
func (t *defaultLogger) Errorf(format string, args ...interface{}) {
	log.Printf("E! "+t.prefix+format, args...)
	for _, f := range t.onError {
		f()
	}
}

// Error logs an error message, patterned after log.Print.
func (t *defaultLogger) Error(args ...interface{}) {
	for _, f := range t.onError {
		f()
	}
	log.Print(append([]interface{}{"E! " + t.prefix}, args...)...)
}

// Debugf logs a debug message, patterned after log.Printf.
func (t *defaultLogger) Debugf(format string, args ...interface{}) {
	log.Printf("D! "+t.prefix+" "+format, args...)
}

// Debug logs a debug message, patterned after log.Print.
func (t *defaultLogger) Debug(args ...interface{}) {
	log.Print(append([]interface{}{"D! " + t.prefix}, args...)...)
}

// Warnf logs a warning message, patterned after log.Printf.
func (t *defaultLogger) Warnf(format string, args ...interface{}) {
	log.Printf("W! "+t.prefix+format, args...)
}

// Warn logs a warning message, patterned after log.Print.
func (t *defaultLogger) Warn(args ...interface{}) {
	log.Print(append([]interface{}{"W! " + t.prefix}, args...)...)
}

// Infof logs an information message, patterned after log.Printf.
func (t *defaultLogger) Infof(format string, args ...interface{}) {
	log.Printf("I! "+t.prefix+format, args...)
}

// Info logs an information message, patterned after log.Print.
func (t *defaultLogger) Info(args ...interface{}) {
	log.Print(append([]interface{}{"I! " + t.prefix}, args...)...)
}

func createDefaultLogger(cfg *Config) (logger, error) {
	log.SetFlags(0)

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
		writer:         wlog.NewWriter(writer),
		internalWriter: writer,
		timezone:       tz,
	}

	log.SetOutput(l)
	return l, nil
}

func init() {
	add("stderr", createDefaultLogger)
	add("file", createDefaultLogger)
}
