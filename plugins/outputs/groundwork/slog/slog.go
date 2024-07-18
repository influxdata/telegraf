package slog

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"

	"github.com/influxdata/telegraf"
)

// NewLogger creates telegraf.Logger adapter for slog.Logger
func NewLogger(l telegraf.Logger) *slog.Logger {
	return slog.New(&TlgHandler{Log: l})
}

// TlgHandler translates slog.Record into telegraf.Logger call
// inspired by https://github.com/golang/example/blob/master/slog-handler-guide/README.md
type TlgHandler struct {
	attrs  []slog.Attr
	groups []string

	once sync.Once

	GroupsFieldName  string
	MessageFieldName string

	Log telegraf.Logger
}

func (h *TlgHandler) Enabled(_ context.Context, level slog.Level) bool {
	l := h.Log.Level()
	switch level {
	case slog.LevelDebug:
		return l >= telegraf.Debug
	case slog.LevelInfo:
		return l >= telegraf.Info
	case slog.LevelWarn:
		return l >= telegraf.Warn
	case slog.LevelError:
		return l >= telegraf.Error
	default:
		return l >= telegraf.Info
	}
}

func (h *TlgHandler) Handle(_ context.Context, r slog.Record) error {
	h.once.Do(func() {
		if h.GroupsFieldName == "" {
			h.GroupsFieldName = "logger"
		}
		if h.MessageFieldName == "" {
			h.MessageFieldName = "message"
		}
	})

	attrs := make([]slog.Attr, 0, 2+len(h.attrs)+r.NumAttrs())
	attrs = append(attrs,
		slog.String(h.MessageFieldName, r.Message),
		slog.String(h.GroupsFieldName, strings.Join(h.groups, ",")),
	)
	for _, attr := range h.attrs {
		if v, ok := attr.Value.Any().(json.RawMessage); ok {
			attrs = append(attrs, slog.String(attr.Key, string(v)))
			continue
		}
		attrs = append(attrs, attr)
	}
	r.Attrs(func(attr slog.Attr) bool {
		if v, ok := attr.Value.Any().(json.RawMessage); ok {
			attrs = append(attrs, slog.String(attr.Key, string(v)))
			return true
		}
		attrs = append(attrs, attr)
		return true
	})

	var handle func(args ...interface{})
	switch r.Level {
	case slog.LevelDebug:
		handle = h.Log.Debug
	case slog.LevelInfo:
		handle = h.Log.Info
	case slog.LevelWarn:
		handle = h.Log.Warn
	case slog.LevelError:
		handle = h.Log.Error
	default:
		handle = h.Log.Info
	}
	handle(attrs)

	return nil
}

func (h *TlgHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	nested := &TlgHandler{GroupsFieldName: h.GroupsFieldName, MessageFieldName: h.MessageFieldName, Log: h.Log}
	nested.attrs = append(nested.attrs, h.attrs...)
	nested.groups = append(nested.groups, h.groups...)
	nested.attrs = append(nested.attrs, attrs...)
	return nested
}

func (h *TlgHandler) WithGroup(name string) slog.Handler {
	nested := &TlgHandler{GroupsFieldName: h.GroupsFieldName, MessageFieldName: h.MessageFieldName, Log: h.Log}
	nested.attrs = append(nested.attrs, h.attrs...)
	nested.groups = append(nested.groups, h.groups...)
	nested.groups = append(nested.groups, name)
	return nested
}
