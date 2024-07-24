//go:build windows

package logger

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/influxdata/telegraf"
	"golang.org/x/sys/windows/svc/eventlog"
)

const (
	eidInfo    = 1
	eidWarning = 2
	eidError   = 3
)

type eventLogger struct {
	eventlog *eventlog.Log
	errlog   *log.Logger
}

func (l *eventLogger) Close() error {
	return l.eventlog.Close()
}

func (l *eventLogger) Print(level telegraf.LogLevel, _ time.Time, prefix string, args ...interface{}) {
	// Skip debug and beyond as they cannot be logged
	if level >= telegraf.Debug {
		return
	}

	msg := level.Indicator() + " " + prefix + fmt.Sprint(args...)

	var err error
	switch level {
	case telegraf.Error:
		err = l.eventlog.Error(eidError, msg)
	case telegraf.Warn:
		err = l.eventlog.Warning(eidWarning, msg)
	case telegraf.Info:
		err = l.eventlog.Info(eidInfo, msg)
	}
	if err != nil {
		l.errlog.Printf("E! Writing log message failed: %v", err)
	}
}

func createEventLogger(cfg *Config) (sink, error) {
	eventLog, err := eventlog.Open(cfg.InstanceName)
	if err != nil {
		return nil, err
	}

	l := &eventLogger{
		eventlog: eventLog,
		errlog:   log.New(os.Stderr, "", 0),
	}

	return l, nil
}

func init() {
	add("eventlog", createEventLogger)
}
