package kafka

import (
	"github.com/Shopify/sarama"

	"github.com/influxdata/telegraf"
)

type Logger struct {
}

// DebugLogger logs messages from sarama at the debug level.
type DebugLogger struct {
	Log telegraf.Logger
}

func (l *DebugLogger) Print(v ...interface{}) {
	args := make([]interface{}, 0, len(v)+1)
	args = append(append(args, "[sarama] "), v...)
	l.Log.Debug(args...)
}

func (l *DebugLogger) Printf(format string, v ...interface{}) {
	l.Log.Debugf("[sarama] "+format, v...)
}

func (l *DebugLogger) Println(v ...interface{}) {
	l.Print(v...)
}

// SetLogger configures a debug logger for kafka (sarama)
func (k *Logger) SetLogger(log telegraf.Logger) {
	sarama.Logger = &DebugLogger{Log: log}
}
