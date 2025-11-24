package heartbeat

import (
	"time"

	"github.com/influxdata/telegraf"
)

func (h *Heartbeat) handleLogEvent(level telegraf.LogLevel, _ time.Time, _ string, _ map[string]interface{}, _ ...interface{}) {
	// Fill the statistics
	h.stats.Lock()
	switch level {
	case telegraf.Error:
		h.stats.logErrors++
	case telegraf.Warn:
		h.stats.logWarnings++
	}
	h.stats.Unlock()
}
