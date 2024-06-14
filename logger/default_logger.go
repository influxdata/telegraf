package logger

import (
	"errors"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/influxdata/telegraf/internal/rotate"
	"github.com/influxdata/wlog"
)

var prefixRegex = regexp.MustCompile("^[DIWE]!")

const (
	LogTargetFile   = "file"
	LogTargetStderr = "stderr"
)

type defaultLogger struct {
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

// newTelegrafWriter returns a logging-wrapped writer.
func newTelegrafWriter(w io.Writer, c Config) (*defaultLogger, error) {
	timezoneName := c.LogWithTimezone
	if strings.EqualFold(timezoneName, "local") {
		timezoneName = "Local"
	}

	tz, err := time.LoadLocation(timezoneName)
	if err != nil {
		return nil, errors.New("error while setting logging timezone: " + err.Error())
	}

	return &defaultLogger{
		writer:         wlog.NewWriter(w),
		internalWriter: w,
		timezone:       tz,
	}, nil
}

func createStderrLogger(cfg Config) (io.WriteCloser, error) {
	return newTelegrafWriter(os.Stderr, cfg)
}

func createFileLogger(cfg Config) (io.WriteCloser, error) {
	if cfg.Logfile == "" {
		return createStderrLogger(cfg)
	}

	writer, err := rotate.NewFileWriter(
		cfg.Logfile,
		cfg.RotationInterval,
		cfg.RotationMaxSize,
		cfg.RotationMaxArchives,
	)
	if err != nil {
		log.Printf("E! Unable to open %s (%s), using stderr", cfg.Logfile, err)
		return createStderrLogger(cfg)
	}

	return newTelegrafWriter(writer, cfg)
}

func init() {
	registerLogger(LogTargetStderr, createStderrLogger)
	registerLogger(LogTargetFile, createFileLogger)
}
