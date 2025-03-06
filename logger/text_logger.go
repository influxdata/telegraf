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

// Keep those constants for backward compatibility even though they are not
// used anywhere. See https://github.com/influxdata/telegraf/pull/15514 for
// more details.
//
// Deprecated: Those constants are unused and deprecated. The removal is
// scheduled for v1.45.0, if you use them please adapt your code!
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
	add("text", createTextLogger)
}
