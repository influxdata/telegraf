package groundwork

import (
	"context"
	"encoding/json"
	"log/slog" //nolint:depguard // Required for wrapping internal logging facility
	"strings"

	"github.com/influxdata/telegraf"
)

// newLogger creates telegraf.Logger adapter for slog.Logger
func newLogger(l telegraf.Logger) *slog.Logger {
	return slog.New(&tlgHandler{Log: l})
}

// tlgHandler translates slog.Record into telegraf.Logger call
// inspired by https://github.com/golang/example/blob/master/slog-handler-guide/README.md
type tlgHandler struct {
	attrs  []slog.Attr
	groups []string

	Log telegraf.Logger
}

// Enabled implements slog.Handler interface
// It interprets errors as errors and everything else as debug.
func (h *tlgHandler) Enabled(_ context.Context, level slog.Level) bool {
	if level == slog.LevelError {
		return h.Log.Level() >= telegraf.Error
	}
	return h.Log.Level() >= telegraf.Debug
}

// Handle implements slog.Handler interface
// It interprets errors as errors and everything else as debug.
func (h *tlgHandler) Handle(_ context.Context, r slog.Record) error {
	attrs := make([]slog.Attr, 0, 2+len(h.attrs)+r.NumAttrs())
	attrs = append(attrs,
		slog.String("logger", strings.Join(h.groups, ",")),
		slog.String("message", r.Message),
	)
	attrs = append(attrs, h.attrs...)

	r.Attrs(func(attr slog.Attr) bool {
		if v, ok := attr.Value.Any().(json.RawMessage); ok {
			attrs = append(attrs, slog.String(attr.Key, string(v)))
			return true
		}
		attrs = append(attrs, attr)
		return true
	})

	if r.Level == slog.LevelError {
		h.Log.Error(attrs)
	} else {
		h.Log.Debug(attrs)
	}

	return nil
}

// WithAttrs implements slog.Handler interface
func (h *tlgHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nested := &tlgHandler{Log: h.Log}
	nested.attrs = append(nested.attrs, h.attrs...)
	nested.groups = append(nested.groups, h.groups...)

	for _, attr := range attrs {
		if v, ok := attr.Value.Any().(json.RawMessage); ok {
			nested.attrs = append(nested.attrs, slog.String(attr.Key, string(v)))
			continue
		}
		nested.attrs = append(nested.attrs, attr)
	}

	return nested
}

// WithGroup implements slog.Handler interface
func (h *tlgHandler) WithGroup(name string) slog.Handler {
	nested := &tlgHandler{Log: h.Log}
	nested.attrs = append(nested.attrs, h.attrs...)
	nested.groups = append(nested.groups, h.groups...)
	nested.groups = append(nested.groups, name)
	return nested
}
