package logger

import (
	"io"
	"log"

	"github.com/influxdata/wlog"
)

// Keep track what is actually set as a log output, because log package doesn't provide a getter.
// It allows closing previous writer if re-set and have possibility to test what is actually set
var actualLogger io.WriteCloser

// SetupLogging configures the logging output.
func SetupLogging(cfg Config) error {
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
	logCreator, ok := loggerRegistry[cfg.LogTarget]
	if !ok {
		log.Printf("E! Unsupported logtarget: %s, using stderr", cfg.LogTarget)
		logCreator = createStderrLogger
	}

	// Create the root logging instance
	logWriter, err := logCreator(cfg)
	if err != nil {
		return err
	}

	// Close the previous logger if possible
	if err := CloseLogging(); err != nil {
		return err
	}

	// Use the new logger and store a reference
	log.SetOutput(logWriter)
	actualLogger = logWriter

	return nil
}

func CloseLogging() error {
	if actualLogger == nil {
		return nil
	}

	return actualLogger.Close()
}
