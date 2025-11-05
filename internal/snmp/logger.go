package snmp

import (
	"fmt"

	"github.com/influxdata/telegraf"
)

type Logger struct {
	gs Connection // Reference to the SNMP connection as the host is not available initially
	telegraf.Logger
}

func (l *Logger) Print(args ...interface{}) {
	message := fmt.Sprint(args...)
	if l.gs != nil && l.gs.Host() != "" {
		message = fmt.Sprintf("agent %s: %s", l.gs.Host(), message)
	}
	l.Trace(message)
}
func (l *Logger) Printf(format string, args ...interface{}) {
	if l.gs != nil && l.gs.Host() != "" {
		format = fmt.Sprintf("agent %s: %s", l.gs.Host(), format)
	}
	l.Tracef(format, args...)
}
