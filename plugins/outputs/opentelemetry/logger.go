package opentelemetry

import (
	"strings"

	"github.com/influxdata/telegraf"
)

type otelLogger struct {
	telegraf.Logger
}

func (l otelLogger) Debug(msg string, kv ...any) {
	format := msg + strings.Repeat(" %s=%q", len(kv)/2)
	l.Logger.Debugf(format, kv...)
}
