package opentelemetry

import (
	"strings"

	"github.com/influxdata/telegraf"
)

type otelLogger struct {
	telegraf.Logger
}

// Debug logs a debug message, patterned after log.Print.
func (l otelLogger) Debug(msg string, kv ...interface{}) {
	format := msg + strings.Repeat(" %s=%q", len(kv)/2)
	l.Logger.Debugf(format, kv...)
}
