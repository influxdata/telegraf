package coralogix

import (
	"strings"

	"github.com/influxdata/telegraf"
)

type logger struct {
	telegraf.Logger
}

func (l logger) Debug(msg string, kv ...interface{}) {
	format := msg + strings.Repeat(" %s=%q", len(kv)/2)
	l.Logger.Debugf(format, kv...)
}
