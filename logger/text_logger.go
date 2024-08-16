package logger

import (
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal/rotate"
)

const (
	LogTargetFile   = "file"
	LogTargetStderr = "stderr"
)

type textLogger struct {
	logger *log.Logger
}

func (l *textLogger) Close() error {
	writer := l.logger.Writer()

	// Close the writer if possible and avoid closing stderr
	if writer == os.Stderr {
		return nil
	}
	if closer, ok := writer.(io.Closer); ok {
		return closer.Close()
	}

	return errors.New("the underlying writer cannot be closed")
}

func (l *textLogger) Print(level telegraf.LogLevel, ts time.Time, prefix string, _ map[string]interface{}, args ...interface{}) {
	msg := append([]interface{}{ts.Format(time.RFC3339), " ", level.Indicator(), " ", prefix}, args...)
	l.logger.Print(msg...)
}

func createTextLogger(cfg *Config) (sink, error) {
	var writer io.Writer = os.Stderr
	if cfg.Logfile != "" {
		w, err := rotate.NewFileWriter(
			cfg.Logfile,
			cfg.RotationInterval,
			cfg.RotationMaxSize,
			cfg.RotationMaxArchives,
		)
		if err != nil {
			return nil, err
		}
		writer = w
	}

	return &textLogger{logger: log.New(writer, "", 0)}, nil
}

func init() {
	add("stderr", func(*Config) (sink, error) {
		msg := "Value %q is deprecated for agent setting %q please use %q instead and leave %q empty!"
		deprecation := "The value will be removed in v1.40.0."
		log.Printf("W! "+msg+" "+deprecation, "stderr", "logtarget", "text", "logfile")

		return &textLogger{logger: log.New(os.Stderr, "", 0)}, nil
	})
	add("file", func(cfg *Config) (sink, error) {
		msg := "Value %q is deprecated for agent setting %q please use %q instead!"
		deprecation := "The value will be removed in v1.40.0."
		log.Printf("W! "+msg+" "+deprecation, "file", "logtarget", "text")

		return createTextLogger(cfg)
	})

	add("text", createTextLogger)
}
