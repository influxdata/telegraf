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

type defaultLogger struct {
	logger *log.Logger
}

func (l *defaultLogger) Close() error {
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

func (l *defaultLogger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

func (l *defaultLogger) Print(level telegraf.LogLevel, ts time.Time, prefix string, args ...interface{}) {
	msg := append([]interface{}{ts.Format(time.RFC3339), " ", level.Indicator(), " ", prefix}, args...)
	l.logger.Print(msg...)
}

func createDefaultLogger(cfg *Config) (sink, error) {
	var writer io.Writer = os.Stderr
	if cfg.LogTarget == "file" && cfg.Logfile != "" {
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

	return &defaultLogger{logger: log.New(writer, "", 0)}, nil
}

func init() {
	add("stderr", createDefaultLogger)
	add("file", createDefaultLogger)
}
