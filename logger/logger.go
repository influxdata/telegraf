package logger

import (
	"errors"
	"io"
	"log"
	"os"
	"regexp"
	"time"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/internal/rotate"
	"github.com/influxdata/wlog"
)

var prefixRegex = regexp.MustCompile("^[DIWE]!")

// newTelegrafWriter returns a logging-wrapped writer.
func newTelegrafWriter(w io.Writer) io.Writer {
	return &telegrafLog{
		writer:         wlog.NewWriter(w),
		internalWriter: w,
	}
}

// LogConfig contains the log configuration settings
type LogConfig struct {
	// will set the log level to DEBUG
	Debug bool
	//will set the log level to ERROR
	Quiet bool
	// will direct the logging output to a file. Empty string is
	// interpreted as stderr. If there is an error opening the file the
	// logger will fallback to stderr
	Logfile string
	// will rotate when current file at the specified time interval
	RotationInterval internal.Duration
	// will rotate when current file size exceeds this parameter.
	RotationMaxSize internal.Size
	// maximum rotated files to keep (older ones will be deleted)
	RotationMaxArchives int
}

type telegrafLog struct {
	writer         io.Writer
	internalWriter io.Writer
}

func (t *telegrafLog) Write(b []byte) (n int, err error) {
	var line []byte
	if !prefixRegex.Match(b) {
		line = append([]byte(time.Now().UTC().Format(time.RFC3339)+" I! "), b...)
	} else {
		line = append([]byte(time.Now().UTC().Format(time.RFC3339)+" "), b...)
	}
	return t.writer.Write(line)
}

func (t *telegrafLog) Close() error {
	closer, isCloser := t.internalWriter.(io.Closer)
	if !isCloser {
		return errors.New("the underlying writer cannot be closed")
	}
	return closer.Close()
}

// SetupLogging configures the logging output.
func SetupLogging(config LogConfig) {
	newLogWriter(config)
}

func newLogWriter(config LogConfig) io.Writer {
	log.SetFlags(0)
	if config.Debug {
		wlog.SetLevel(wlog.DEBUG)
	}
	if config.Quiet {
		wlog.SetLevel(wlog.ERROR)
	}

	var writer io.Writer
	if config.Logfile != "" {
		var err error
		if writer, err = rotate.NewFileWriter(config.Logfile, config.RotationInterval.Duration, config.RotationMaxSize.Size, config.RotationMaxArchives); err != nil {
			log.Printf("E! Unable to open %s (%s), using stderr", config.Logfile, err)
			writer = os.Stderr
		}
	} else {
		writer = os.Stderr
	}

	telegrafLog := newTelegrafWriter(writer)
	log.SetOutput(telegrafLog)
	return telegrafLog
}
