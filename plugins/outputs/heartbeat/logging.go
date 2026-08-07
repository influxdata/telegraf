package heartbeat

import (
	"cmp"
	"container/ring"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/influxdata/telegraf"
)

type LogsConfig struct {
	Limit    uint64 `toml:"limit"`
	LogLevel string `toml:"level"`

	level telegraf.LogLevel
}

type logEvent struct {
	timestamp  time.Time
	level      telegraf.LogLevel
	source     string
	attributes map[string]interface{}
	msg        string
}

func (h *Heartbeat) handleLogEvent(level telegraf.LogLevel, ts time.Time, source string, attr map[string]interface{}, args ...interface{}) {
	// Fill the statistics
	h.stats.Lock()
	switch level {
	case telegraf.Error:
		h.stats.logErrors++
	case telegraf.Warn:
		h.stats.logWarnings++
	}
	h.stats.Unlock()

	// Only save events if the logging configuration requests us to do so
	if !h.Logs.level.Includes(level) {
		return
	}

	// Add the event
	h.Lock()
	h.logEvents = append(h.logEvents, &logEvent{
		timestamp:  ts,
		level:      level,
		source:     source,
		attributes: maps.Clone(attr),
		msg:        fmt.Sprint(args...),
	})
	h.Unlock()
}

func getLogEntriesUnlimited(events []*logEvent) []logEntry {
	entries := make([]logEntry, 0, len(events))
	for _, e := range events {
		entries = append(entries, logEntry{
			Timestamp:  e.timestamp.Format(time.RFC3339Nano),
			Level:      e.level.String(),
			Source:     e.source,
			Attributes: e.attributes,
			Message:    e.msg,
		})
	}

	return entries
}

func getLogEntriesLimited(events []*logEvent, limit int) []logEntry {
	// Collect all events per log level for filtering
	tracker := make(map[telegraf.LogLevel]*ring.Ring)
	for i, e := range events {
		if tracker[e.level] == nil {
			tracker[e.level] = ring.New(limit)
		} else {
			tracker[e.level] = tracker[e.level].Next()
		}
		tracker[e.level].Value = logEntry{
			Timestamp:  e.timestamp.Format(time.RFC3339Nano),
			Level:      e.level.String(),
			Source:     e.source,
			Attributes: e.attributes,
			Message:    e.msg,
			index:      i,
		}
	}

	// Define log-level with priorities
	loglevels := []telegraf.LogLevel{
		telegraf.Error,
		telegraf.Warn,
		telegraf.Info,
		telegraf.Debug,
		telegraf.Trace,
	}

	// Unroll the ringbuffers until the limit is reached. Start from the most
	// severe log-level to the least severe one.
	var count int
	entries := make([]logEntry, 0, limit)
	for _, level := range loglevels {
		for r := tracker[level]; r.Value != nil && count < limit; r = r.Prev() {
			count++
			entries = append(entries, r.Value.(logEntry))
		}

		if count >= limit {
			break
		}
	}

	// Restore the temporal order of the log entries
	slices.SortFunc(entries, func(a, b logEntry) int { return cmp.Compare(a.index, b.index) })

	return entries
}
