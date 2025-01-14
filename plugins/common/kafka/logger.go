package kafka

import (
	"sync"

	"github.com/IBM/sarama"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/logger"
)

var (
	log  = logger.New("sarama", "", "")
	once sync.Once
)

type debugLogger struct{}

func (*debugLogger) Print(v ...interface{}) {
	log.Trace(v...)
}

func (*debugLogger) Printf(format string, v ...interface{}) {
	log.Tracef(format, v...)
}

func (l *debugLogger) Println(v ...interface{}) {
	l.Print(v...)
}

// SetLogger configures a debug logger for kafka (sarama)
func SetLogger(level telegraf.LogLevel) {
	// Set-up the sarama logger only once
	once.Do(func() {
		sarama.Logger = &debugLogger{}
	})
	// Increase the log-level if needed.
	if !log.Level().Includes(level) {
		log.SetLevel(level)
	}
}
