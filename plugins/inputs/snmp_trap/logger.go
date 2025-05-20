package snmp_trap

import "github.com/influxdata/telegraf"

type logger struct {
	telegraf.Logger
}

// Printf formats and writes the given string to the standard output.
func (l logger) Printf(format string, args ...interface{}) {
	l.Tracef(format, args...)
}

// Print writes the given string to the standard output.
func (l logger) Print(args ...interface{}) {
	l.Trace(args...)
}
