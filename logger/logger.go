package logger

import (
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
)

// Central handler for the logs used by the logger to actually output the logs.
// This is necessary to be able to dynamically switch the sink even though
// plugins already instantiated a logger _before_ the final sink is set up.
var (
	instance *handler  // handler for the actual output
	once     sync.Once // once token to initialize the handler only once
)

// sink interface that has to be implemented by a logging sink
type sink interface {
	Print(telegraf.LogLevel, time.Time, string, ...interface{})
}

// Attr represents an attribute appended to structured logging
type Attr struct {
	Key   string
	Value interface{}
}

// logger is the actual implementation of the telegraf logger interface
type logger struct {
	level    *telegraf.LogLevel
	category string
	name     string
	alias    string
	suffix   string

	prefix string

	onError []func()
}

// New creates a new logging instance to be used in models
func New(category, name, alias string) *logger {
	l := &logger{
		category: category,
		name:     name,
		alias:    alias,
	}
	l.formatPrefix()

	return l
}

// SubLogger creates a new logger with the given name added as suffix
func (l *logger) SubLogger(name string) telegraf.Logger {
	suffix := l.suffix
	if suffix != "" && name != "" {
		suffix += "."
	}
	suffix += name

	nl := &logger{
		level:    l.level,
		category: l.category,
		name:     l.name,
		alias:    l.alias,
		suffix:   suffix,
	}
	nl.formatPrefix()

	return nl
}

func (l *logger) formatPrefix() {
	l.prefix = l.category

	if l.prefix != "" && l.name != "" {
		l.prefix += "."
	}
	l.prefix += l.name

	if l.prefix != "" && l.alias != "" {
		l.prefix += "::"
	}
	l.prefix += l.alias

	if l.suffix != "" {
		l.prefix += "(" + l.suffix + ")"
	}

	if l.prefix != "" {
		l.prefix = "[" + l.prefix + "] "
	}
}

// Level returns the current log-level of the logger
func (l *logger) Level() telegraf.LogLevel {
	if l.level != nil {
		return *l.level
	}
	return instance.level
}

// Register a callback triggered when errors are about to be written to the log
func (l *logger) RegisterErrorCallback(f func()) {
	l.onError = append(l.onError, f)
}

// Error logging including callbacks
func (l *logger) Errorf(format string, args ...interface{}) {
	l.Error(fmt.Sprintf(format, args...))
}

func (l *logger) Error(args ...interface{}) {
	l.Print(telegraf.Error, time.Now(), args...)
	for _, f := range l.onError {
		f()
	}
}

// Warning logging
func (l *logger) Warnf(format string, args ...interface{}) {
	l.Warn(fmt.Sprintf(format, args...))
}

func (l *logger) Warn(args ...interface{}) {
	l.Print(telegraf.Warn, time.Now(), args...)
}

// Info logging
func (l *logger) Infof(format string, args ...interface{}) {
	l.Info(fmt.Sprintf(format, args...))
}

func (l *logger) Info(args ...interface{}) {
	l.Print(telegraf.Info, time.Now(), args...)
}

// Debug logging, this is suppressed on console
func (l *logger) Debugf(format string, args ...interface{}) {
	l.Debug(fmt.Sprintf(format, args...))
}

func (l *logger) Debug(args ...interface{}) {
	l.Print(telegraf.Debug, time.Now(), args...)
}

func (l *logger) Print(level telegraf.LogLevel, ts time.Time, args ...interface{}) {
	// Check if we are in early logging state and store the message in this case
	if instance.impl == nil {
		instance.add(level, ts, l.prefix, args...)
	}

	// Skip all messages with insufficient log-levels
	if l.level != nil && !l.level.Includes(level) || l.level == nil && !instance.level.Includes(level) {
		return
	}
	if instance.impl != nil {
		instance.impl.Print(level, ts.In(instance.timezone), l.prefix, args...)
	} else {
		msg := append([]interface{}{ts.In(instance.timezone).Format(time.RFC3339), " ", level.Indicator(), " ", l.prefix}, args...)
		instance.earlysink.Print(msg...)
	}
}

type Config struct {
	// will set the log level to DEBUG
	Debug bool
	// will set the log level to ERROR
	Quiet bool
	//stderr, stdout, file or eventlog (Windows only)
	LogTarget string
	// will direct the logging output to a file. Empty string is
	// interpreted as stderr. If there is an error opening the file the
	// logger will fall back to stderr
	Logfile string
	// will rotate when current file at the specified time interval
	RotationInterval time.Duration
	// will rotate when current file size exceeds this parameter.
	RotationMaxSize int64
	// maximum rotated files to keep (older ones will be deleted)
	RotationMaxArchives int
	// pick a timezone to use when logging. or type 'local' for local time.
	LogWithTimezone string
	// Logger instance name
	InstanceName string

	// internal  log-level
	logLevel telegraf.LogLevel
}

// SetupLogging configures the logging output.
func SetupLogging(cfg *Config) error {
	if cfg.Debug {
		cfg.logLevel = telegraf.Debug
	}
	if cfg.Quiet {
		cfg.logLevel = telegraf.Error
	}
	if !cfg.Debug && !cfg.Quiet {
		cfg.logLevel = telegraf.Info
	}

	if cfg.InstanceName == "" {
		cfg.InstanceName = "telegraf"
	}

	if cfg.LogTarget == "" || cfg.LogTarget == "file" && cfg.Logfile == "" {
		cfg.LogTarget = "stderr"
	}

	// Get configured timezone
	timezoneName := cfg.LogWithTimezone
	if strings.EqualFold(timezoneName, "local") {
		timezoneName = "Local"
	}
	tz, err := time.LoadLocation(timezoneName)
	if err != nil {
		return fmt.Errorf("setting logging timezone failed: %w", err)
	}

	// Get the logging factory
	creator, ok := registry[cfg.LogTarget]
	if !ok {
		return fmt.Errorf("unsupported log target: %s, using stderr", cfg.LogTarget)
	}

	// Create the root logging instance
	l, err := creator(cfg)
	if err != nil {
		return err
	}

	// Close the previous logger if possible
	if err := CloseLogging(); err != nil {
		return err
	}

	// Update the logging instance
	instance.switchSink(l, cfg.logLevel, tz, cfg.LogTarget == "stderr")

	return nil
}

func RedirectLogging(w io.Writer) {
	instance = redirectHandler(w)
}

func CloseLogging() error {
	return instance.close()
}

func init() {
	once.Do(func() {
		// Create a special logging instance that additionally buffers all
		// messages logged before the final logger is up.
		instance = defaultHandler()

		// Redirect the standard logger output to our logger instance
		log.SetFlags(0)
		log.SetOutput(&stdlogRedirector{})
	})
}
