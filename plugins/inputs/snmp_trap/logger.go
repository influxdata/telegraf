package snmp_trap

import "github.com/influxdata/telegraf"

type logger struct {
	telegraf.Logger
}

func (l logger) Printf(format string, args ...interface{}) {
	l.Debugf(format, args...)
}

func (l logger) Print(args ...interface{}) {
	l.Debug(args...)
}
