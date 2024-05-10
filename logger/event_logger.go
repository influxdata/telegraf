//go:build windows

package logger

import (
	"io"
	"log"
	"strings"

	"github.com/influxdata/wlog"
	"golang.org/x/sys/windows/svc/eventlog"
)

const (
	eidInfo    = 1
	eidWarning = 2
	eidError   = 3
)

type eventWriter struct {
	logger *eventlog.Log
}

func (w *eventWriter) Write(b []byte) (int, error) {
	loc := prefixRegex.FindIndex(b)
	n := len(b)
	if loc == nil {
		return n, w.logger.Info(1, string(b))
	}

	//skip empty log messages
	if n > 2 {
		line := strings.Trim(string(b[loc[1]:]), " \t\r\n")
		switch rune(b[loc[0]]) {
		case 'I':
			return n, w.logger.Info(eidInfo, line)
		case 'W':
			return n, w.logger.Warning(eidWarning, line)
		case 'E':
			return n, w.logger.Error(eidError, line)
		}
	}

	return n, nil
}

type eventLogger struct {
	writer   io.Writer
	eventlog *eventlog.Log
}

func (e *eventLogger) Write(b []byte) (int, error) {
	return e.writer.Write(b)
}

func (e *eventLogger) Close() error {
	return e.eventlog.Close()
}

func createEventLogger(name string) creator {
	return func(Config) (io.WriteCloser, error) {
		eventLog, err := eventlog.Open(name)
		if err != nil {
			log.Printf("E! An error occurred while initializing an event logger. %s", err)
			return nil, err
		}

		writer := wlog.NewWriter(&eventWriter{logger: eventLog})
		return &eventLogger{
			writer:   writer,
			eventlog: eventLog,
		}, nil
	}
}

func RegisterEventLogger(name string) error {
	registerLogger("eventlog", createEventLogger(name))
	return nil
}
