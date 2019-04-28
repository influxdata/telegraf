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
		writer: wlog.NewWriter(w),
	}
}

type telegrafLog struct {
	writer io.Writer
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
	closer, isCloser := t.writer.(io.Closer)
	if !isCloser {
		return errors.New("the underlying writer cannot be closed")
	}
	return closer.Close()
}

// SetupLogging configures the logging output.
//   debug                  will set the log level to DEBUG
//   quiet                  will set the log level to ERROR
//   logfile                will direct the logging output to a file. Empty string is
//                          interpreted as stderr. If there is an error opening the file the
//                          logger will fallback to stderr.
//   logRotationInterval    will rotate when current file at the specified time interval.
//   logRotationMaxSize     will rotate when current file size exceeds this parameter.
//   logRotationMaxArchives maximum rotated files to keep (older ones will be deleted)
func SetupLogging(debug, quiet bool, logfile string, logRotationInterval time.Duration, logRotationMaxSize internal.Size, logRotationMaxArchives int) {
	setupLoggingAndReturnWriter(debug, quiet, logfile, logRotationInterval, logRotationMaxSize, logRotationMaxArchives)
}

func setupLoggingAndReturnWriter(debug, quiet bool, logfile string, logRotationInterval time.Duration, logRotationMaxSize internal.Size,
	logRotationMaxArchives int) io.Writer {
	log.SetFlags(0)
	if debug {
		wlog.SetLevel(wlog.DEBUG)
	}
	if quiet {
		wlog.SetLevel(wlog.ERROR)
	}

	var writer io.Writer
	if logfile != "" {
		var err error
		if writer, err = rotate.NewFileWriter(logfile, logRotationInterval, logRotationMaxSize.Size, logRotationMaxArchives); err != nil {
			log.Printf("E! Unable to open %s (%s), using stderr", logfile, err)
			writer = os.Stderr
		}
	} else {
		writer = os.Stderr
	}

	telegrafLog := newTelegrafWriter(writer)
	log.SetOutput(telegrafLog)
	return telegrafLog
}
