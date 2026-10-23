package mqtt

import (
	"github.com/influxdata/telegraf"
)

type mqttLogger struct {
	telegraf.Logger
}

func (l mqttLogger) Printf(fmt string, args ...any) {
	l.Logger.Debugf(fmt, args...)
}

func (l mqttLogger) Println(args ...any) {
	l.Logger.Debug(args...)
}
