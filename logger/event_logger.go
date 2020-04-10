package logger

import (
	"io"
	"strings"

	"github.com/influxdata/wlog"
	"github.com/kardianos/service"
)

const (
	LogTargetEventlog = "eventlog"
)

type eventLogger struct {
	logger service.Logger
}

func (t *eventLogger) Write(b []byte) (n int, err error) {
	loc := prefixRegex.FindIndex(b)
	n = len(b)
	if loc == nil {
		err = t.logger.Info(b)
	} else if n > 2 { //skip empty log messages
		line := strings.Trim(string(b[loc[1]:]), " \t\r\n")
		switch rune(b[loc[0]]) {
		case 'I':
			err = t.logger.Info(line)
		case 'W':
			err = t.logger.Warning(line)
		case 'E':
			err = t.logger.Error(line)
		}
	}

	return
}

type eventLoggerCreator struct {
	serviceLogger service.Logger
}

func (e *eventLoggerCreator) CreateLogger(config LogConfig) (io.Writer, error) {
	return wlog.NewWriter(&eventLogger{logger: e.serviceLogger}), nil
}

func RegisterEventLogger(serviceLogger service.Logger) {
	registerLogger(LogTargetEventlog, &eventLoggerCreator{serviceLogger: serviceLogger})
}
