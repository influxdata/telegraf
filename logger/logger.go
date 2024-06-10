package logger

import (
	"fmt"
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/wlog"
)

type logger interface {
	telegraf.Logger
	New(category, name, alias string) telegraf.Logger
	Close() error
}

type Config struct {
	// will set the log level to DEBUG
	Debug bool
	//will set the log level to ERROR
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
}

// Keep track what is actually set as a log output, because log package doesn't provide a getter.
// It allows closing previous writer if re-set and have possibility to test what is actually set
var instance logger

// SetupLogging configures the logging output.
func SetupLogging(cfg *Config) error {
	log.SetFlags(0)
	if cfg.Debug {
		wlog.SetLevel(wlog.DEBUG)
	}
	if cfg.Quiet {
		wlog.SetLevel(wlog.ERROR)
	}
	if !cfg.Debug && !cfg.Quiet {
		wlog.SetLevel(wlog.INFO)
	}

	if cfg.LogTarget == "" {
		cfg.LogTarget = LogTargetStderr
	}

	// Get the logging factory
	creator, ok := registry[cfg.LogTarget]
	if !ok {
		return fmt.Errorf("unsupported logtarget: %s, using stderr", cfg.LogTarget)
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

	// Use the new logger and store a reference
	instance = l

	return nil
}

func NewLogger(category, name, alias string) telegraf.Logger {
	if instance == nil {
		SetupLogging(&Config{})
	}
	return instance.New(category, name, alias)
}

func CloseLogging() error {
	if instance != nil {
		return instance.Close()
	}
	return nil
}
