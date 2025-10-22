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

type logEvent struct {
	timestamp  time.Time
	level      telegraf.LogLevel
	source     string
	attributes map[string]interface{}
	msg        string
}

func (h *Heartbeat) handleLogEvent(level telegraf.LogLevel, ts time.Time, source string, attr map[string]interface{}, args ...interface{}) {
	// Fill the statistics
	switch level {
	case telegraf.Error:
		h.logErrors.Add(1)
	case telegraf.Warn:
		h.logWarnings.Add(1)
	}

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

func (h *Heartbeat) getLogEntriesUnlimited() []logEntry {
	h.Lock()
	defer h.Unlock()

	entries := make([]logEntry, 0, len(h.logEvents))
	for _, e := range h.logEvents {
		entries = append(entries, logEntry{
			Timestamp:  e.timestamp.Format(time.RFC3339Nano),
			Level:      e.level.String(),
			Source:     e.source,
			Attributes: e.attributes,
			Messsage:   e.msg,
		})
	}
	clear(h.logEvents)
	return entries
}

func (h *Heartbeat) getLogEntriesLimited() []logEntry {
	h.Lock()
	defer h.Unlock()

	limit := int(h.Logs.Limit)

	// Collect all events per log level for filtering
	tracker := make(map[telegraf.LogLevel]*ring.Ring)
	for i, e := range h.logEvents {
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
			Messsage:   e.msg,
			index:      i,
		}
	}
	clear(h.logEvents)

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
	var count uint64
	entries := make([]logEntry, 0, limit)
	for _, level := range loglevels {
		for r := tracker[level]; r.Value != nil && count < h.Logs.Limit; r = r.Prev() {
			count++
			entries = append(entries, r.Value.(logEntry))
		}

		if count >= h.Logs.Limit {
			break
		}
	}

	// Restore the temporal order of the log entries
	slices.SortFunc(entries, func(a, b logEntry) int { return cmp.Compare(a.index, b.index) })

	return entries
}
