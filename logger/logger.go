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
	RotationInterval internal.Duration
	// will rotate when current file size exceeds this parameter.
	RotationMaxSize internal.Size
	// maximum rotated files to keep (older ones will be deleted)
	RotationMaxArchives int
}

type LoggerCreator interface {
	CreateLogger(config LogConfig) (io.Writer, error)
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
	var stdErrWriter io.Writer
	stdErrWriter = os.Stderr
	// avoid closing stderr
	if t.internalWriter != stdErrWriter {
		closer, isCloser := t.internalWriter.(io.Closer)
		if !isCloser {
			return errors.New("the underlying writer cannot be closed")
		}
		return closer.Close()
	}
	return nil
}

// newTelegrafWriter returns a logging-wrapped writer.
func newTelegrafWriter(w io.Writer) io.Writer {
	return &telegrafLog{
		writer:         wlog.NewWriter(w),
		internalWriter: w,
	}
}

// SetupLogging configures the logging output.
func SetupLogging(config LogConfig) {
	newLogWriter(config)
}

type telegrafLogCreator struct {
}

func (t *telegrafLogCreator) CreateLogger(config LogConfig) (io.Writer, error) {
	var writer, defaultWriter io.Writer
	defaultWriter = os.Stderr

	switch config.LogTarget {
	case LogTargetFile:
		if config.Logfile != "" {
			var err error
			if writer, err = rotate.NewFileWriter(config.Logfile, config.RotationInterval.Duration, config.RotationMaxSize.Size, config.RotationMaxArchives); err != nil {
				log.Printf("E! Unable to open %s (%s), using stderr", config.Logfile, err)
				writer = defaultWriter
			}
		} else {
			writer = defaultWriter
		}
	case LogTargetStderr, "":
		writer = defaultWriter
	default:
		log.Printf("E! Unsupported logtarget: %s, using stderr", config.LogTarget)
		writer = defaultWriter
	}

	return newTelegrafWriter(writer), nil
}

// Keep track what is actually set as a log output, because log package doesn't provide a getter.
// It allows closing previous writer if re-set and have possibility to test what is actually set
var actualLogger io.Writer

func newLogWriter(config LogConfig) io.Writer {
	log.SetFlags(0)
	if config.Debug {
		wlog.SetLevel(wlog.DEBUG)
	}
	if config.Quiet {
		wlog.SetLevel(wlog.ERROR)
	}
	if !config.Debug && !config.Quiet {
		wlog.SetLevel(wlog.INFO)
	}
	var logWriter io.Writer
	if logCreator, ok := loggerRegistry[config.LogTarget]; ok {
		logWriter, _ = logCreator.CreateLogger(config)
	}
	if logWriter == nil {
		logWriter, _ = (&telegrafLogCreator{}).CreateLogger(config)
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
