package snmp_trap

import "github.com/influxdata/telegraf"

type logger struct {
	telegraf.Logger
}

func (l logger) Printf(format string, args ...interface{}) {
	l.Tracef(format, args...)
}

func (l logger) Print(args ...interface{}) {
	l.Trace(args...)
}
