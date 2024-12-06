package kinesis_consumer

import (
	"context"
	"log/slog" //nolint:depguard // required to create a wrapper for using it in the library
	"strings"

	"github.com/aws/smithy-go/logging"
	"github.com/influxdata/telegraf"
)

type telegrafLoggerWrapper struct {
	telegraf.Logger
}

func (t *telegrafLoggerWrapper) Log(args ...interface{}) {
	t.Trace(args...)
}

func (t *telegrafLoggerWrapper) Logf(classification logging.Classification, format string, v ...interface{}) {
	switch classification {
	case logging.Debug:
		format = "DEBUG " + format
	case logging.Warn:
		format = "WARN" + format
	default:
		format = "INFO " + format
	}
	t.Logger.Tracef(format, v...)
}

type logger struct {
	log telegraf.Logger

	name  string
	attrs []slog.Attr
}

func newLogger(l telegraf.Logger) *slog.Logger {
	return slog.New(&logger{log: l})
}

func (*logger) Enabled(context.Context, slog.Level) bool {
	return true
}

func (l *logger) Handle(_ context.Context, r slog.Record) error {
	var msg string
	if l.name != "" {
		msg = "[" + l.name + "] "
	}
	msg += r.Level.String() + " - " + r.Message

	attrs := make([]string, 0, len(l.attrs))
	for _, a := range l.attrs {
		attrs = append(attrs, a.Key+": "+a.Value.String())
	}
	msg += "{" + strings.Join(attrs, ",") + "}"

	l.log.Trace(msg)

	return nil
}

func (l *logger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &logger{
		log:   l.log,
		name:  l.name,
		attrs: append(l.attrs, attrs...),
	}
}

func (l *logger) WithGroup(name string) slog.Handler {
	return &logger{
		log:   l.log,
		name:  strings.Trim(l.name+"."+name, "."),
		attrs: append(make([]slog.Attr, 0), l.attrs...),
	}
}
