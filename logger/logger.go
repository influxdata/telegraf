package logger

import (
	"io"
	"log"
	"os"
	"regexp"
	"time"

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

// SetupLogging configures the logging output.
//   debug   will set the log level to DEBUG
//   quiet   will set the log level to ERROR
//   logfile will direct the logging output to a file. Empty string is
//           interpreted as stderr. If there is an error opening the file the
//           logger will fallback to stderr.
func SetupLogging(debug, quiet bool, logfile string) {
	log.SetFlags(0)
	if debug {
		wlog.SetLevel(wlog.DEBUG)
	}
	if quiet {
		wlog.SetLevel(wlog.ERROR)
	}

	var oFile *os.File
	if logfile != "" {
		if _, err := os.Stat(logfile); os.IsNotExist(err) {
			if oFile, err = os.Create(logfile); err != nil {
				log.Printf("E! Unable to create %s (%s), using stderr", logfile, err)
				oFile = os.Stderr
			}
		} else {
			if oFile, err = os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY, os.ModeAppend); err != nil {
				log.Printf("E! Unable to append to %s (%s), using stderr", logfile, err)
				oFile = os.Stderr
			}
		}
	} else {
		oFile = os.Stderr
	}

	log.SetOutput(newTelegrafWriter(oFile))
}
