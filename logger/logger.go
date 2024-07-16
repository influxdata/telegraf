package logger

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/influxdata/telegraf"
)

var prefixRegex = regexp.MustCompile("^[DIWE]!")

type logger interface {
	telegraf.Logger
	New(tag string) telegraf.Logger
	Print(level telegraf.LogLevel, ts time.Time, args ...interface{})
	Close() error
}

type redirectable interface {
	SetOutput(io.Writer)
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

// Keep track what is actually set as a log output, because log package doesn't provide a getter.
// It allows closing previous writer if re-set and have possibility to test what is actually set
var instance logger
var once sync.Once

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

	if cfg.LogTarget == "" {
		cfg.LogTarget = "stderr"
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

	// Use the new logger and store a reference, transfer early logs if any
	if early, ok := instance.(*earlyLogger); cfg.LogTarget != "stderr" && ok {
		early.buffer.Lock()
		current := early.buffer.entries.Front()
		for current != nil {
			e := current.Value.(*entry)
			l.Print(e.level, e.timestamp, e.args...)
			next := current.Next()
			early.buffer.entries.Remove(current)
			current = next
		}
		early.buffer.Unlock()
	}
	instance = l

	return nil
}

func NewLogger(category, name, alias string) telegraf.Logger {
	prefix := category
	if name != "" {
		if prefix != "" {
			prefix += "." + name
		} else {
			prefix = name
		}
	}
	if alias != "" {
		if prefix != "" {
			prefix += "::" + alias
		} else {
			prefix = alias
		}
	}

	return instance.New(prefix)
}

func RedirectLogging(w io.Writer) {
	if e, ok := instance.(redirectable); ok {
		e.SetOutput(w)
	}
}

func CloseLogging() error {
	if instance != nil {
		return instance.Close()
	}
	return nil
}

func init() {
	once.Do(func() {
		// Create a special logging instance that additionally buffers all
		// messages logged before the final logger is up.
		instance = createEarlyLogger()

		// Redirect the standard logger output to our logger instance
		log.SetFlags(0)
		log.SetOutput(&stdlogRedirector{})
	})
}
