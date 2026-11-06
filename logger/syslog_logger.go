//go:build !windows

package logger

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
	"time"

	"github.com/influxdata/telegraf"
)

// Network and address of the syslog server. Empty values connect to the local
// syslog daemon. Those variables are only modified in unit tests to redirect
// the output.
var (
	syslogNetwork string
	syslogAddress string
)

type syslogLogger struct {
	writer *syslog.Writer
	errlog *log.Logger
}

func (l *syslogLogger) Close() error {
	if l.writer == nil {
		return nil
	}
	if err := l.writer.Close(); err != nil {
		return err
	}
	l.writer = nil
	return nil
}

func (l *syslogLogger) Print(level telegraf.LogLevel, _ time.Time, prefix string, _ map[string]any, args ...any) {
	// Messages cannot be written after closing the logger
	if l.writer == nil {
		return
	}

	// Timestamps are added by syslog itself so we do not need to output them.
	msg := prefix + fmt.Sprint(args...)

	var err error
	switch level {
	case telegraf.Error:
		err = l.writer.Err(msg)
	case telegraf.Warn:
		err = l.writer.Warning(msg)
	case telegraf.Info:
		err = l.writer.Info(msg)
	case telegraf.Debug, telegraf.Trace:
		err = l.writer.Debug(msg)
	default:
		return
	}
	if err != nil {
		l.errlog.Printf("E! Writing log message failed: %v", err)
	}
}

func createSyslogLogger(cfg *Config) (sink, error) {
	w, err := syslog.Dial(syslogNetwork, syslogAddress, syslog.LOG_DAEMON|syslog.LOG_INFO, cfg.InstanceName)
	if err != nil {
		return nil, fmt.Errorf("connecting to syslog failed: %w", err)
	}

	l := &syslogLogger{
		writer: w,
		errlog: log.New(os.Stderr, "", 0),
	}

	return l, nil
}

func init() {
	add("syslog", createSyslogLogger)
}
