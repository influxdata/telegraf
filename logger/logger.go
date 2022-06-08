package logger

import (
	"errors"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal/rotate"
	"github.com/influxdata/wlog"
)

var prefixRegex = regexp.MustCompile("^[DIWE]!")

const (
	LogTargetFile   = "file"
	LogTargetStderr = "stderr"
)

// LogConfig contains the log configuration settings
type LogConfig struct {
	// will set the log level to DEBUG
	Debug bool
	//will set the log level to ERROR
	Quiet bool
	//stderr, stdout, file or eventlog (Windows only)
	LogTarget string
	// will direct the logging output to a file. Empty string is
	// interpreted as stderr. If there is an error opening the file the
	// logger will fallback to stderr
	Logfile string
	// will rotate when current file at the specified time interval
	RotationInterval config.Duration
	// will rotate when current file size exceeds this parameter.
	RotationMaxSize config.Size
	// maximum rotated files to keep (older ones will be deleted)
	RotationMaxArchives int
	// pick a timezone to use when logging. or type 'local' for local time.
	LogWithTimezone string
}

type LoggerCreator interface {
	CreateLogger(cfg LogConfig) (io.Writer, error)
}

var loggerRegistry map[string]LoggerCreator

func registerLogger(name string, loggerCreator LoggerCreator) {
	if loggerRegistry == nil {
		loggerRegistry = make(map[string]LoggerCreator)
	}
	loggerRegistry[name] = loggerCreator
}

type telegrafLog struct {
	writer         io.Writer
	internalWriter io.Writer
	timezone       *time.Location
}

func (t *telegrafLog) Write(b []byte) (n int, err error) {
	var line []byte
	timeToPrint := time.Now().In(t.timezone)

	if !prefixRegex.Match(b) {
		line = append([]byte(timeToPrint.Format(time.RFC3339)+" I! "), b...)
	} else {
		line = append([]byte(timeToPrint.Format(time.RFC3339)+" "), b...)
	}

	return t.writer.Write(line)
}

func (t *telegrafLog) Close() error {
	stdErrWriter := os.Stderr
	// avoid closing stderr
	if t.internalWriter == stdErrWriter {
		return nil
	}

	closer, isCloser := t.internalWriter.(io.Closer)
	if !isCloser {
		return errors.New("the underlying writer cannot be closed")
	}
	return closer.Close()
}

// newTelegrafWriter returns a logging-wrapped writer.
func newTelegrafWriter(w io.Writer, c LogConfig) (io.Writer, error) {
	timezoneName := c.LogWithTimezone

	if strings.ToLower(timezoneName) == "local" {
		timezoneName = "Local"
	}

	tz, err := time.LoadLocation(timezoneName)
	if err != nil {
		return nil, errors.New("error while setting logging timezone: " + err.Error())
	}

	return &telegrafLog{
		writer:         wlog.NewWriter(w),
		internalWriter: w,
		timezone:       tz,
	}, nil
}

// SetupLogging configures the logging output.
func SetupLogging(cfg LogConfig) {
	newLogWriter(cfg)
}

type telegrafLogCreator struct {
}

func (t *telegrafLogCreator) CreateLogger(cfg LogConfig) (io.Writer, error) {
	var writer, defaultWriter io.Writer
	defaultWriter = os.Stderr

	switch cfg.LogTarget {
	case LogTargetFile:
		if cfg.Logfile != "" {
			var err error
			if writer, err = rotate.NewFileWriter(cfg.Logfile, time.Duration(cfg.RotationInterval), int64(cfg.RotationMaxSize), cfg.RotationMaxArchives); err != nil {
				log.Printf("E! Unable to open %s (%s), using stderr", cfg.Logfile, err)
				writer = defaultWriter
			}
		} else {
			writer = defaultWriter
		}
	case LogTargetStderr, "":
		writer = defaultWriter
	default:
		log.Printf("E! Unsupported logtarget: %s, using stderr", cfg.LogTarget)
		writer = defaultWriter
	}

	return newTelegrafWriter(writer, cfg)
}

// Keep track what is actually set as a log output, because log package doesn't provide a getter.
// It allows closing previous writer if re-set and have possibility to test what is actually set
var actualLogger io.Writer

func newLogWriter(cfg LogConfig) io.Writer {
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
	var logWriter io.Writer
	if logCreator, ok := loggerRegistry[cfg.LogTarget]; ok {
		logWriter, _ = logCreator.CreateLogger(cfg)
	}
	if logWriter == nil {
		logWriter, _ = (&telegrafLogCreator{}).CreateLogger(cfg)
	}

	if closer, isCloser := actualLogger.(io.Closer); isCloser {
		closer.Close()
	}
	log.SetOutput(logWriter)
	actualLogger = logWriter

	return logWriter
}

func init() {
	tlc := &telegrafLogCreator{}
	registerLogger("", tlc)
	registerLogger(LogTargetStderr, tlc)
	registerLogger(LogTargetFile, tlc)
}
