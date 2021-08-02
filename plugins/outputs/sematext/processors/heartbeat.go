package processors

import (
	"sync"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

const (
	oneMinuteSeconds = int64(60)
	oneHourMinutes   = int64(60)
	oneDaySeconds    = 24 * oneHourMinutes * oneMinuteSeconds
	oneDayMinutes    = 24 * oneHourMinutes
)

// Heartbeat is a batch processor that injects heartbeat metric as necessary (once per minute). It stores info about
// already injected heartbeats (one per minute) into injectedMinutes field. It will clear this map once a day to avoid
// it to grow too big (field mapResetDay keeps the record of the "day" for which injectedMinutes contains the data)
type Heartbeat struct {
	injectedMinutes map[int64]bool
	mapResetDay     int64
	lock            sync.Mutex
}

func NewHeartbeat() BatchProcessor {
	return &Heartbeat{
		injectedMinutes: make(map[int64]bool),
	}
}

// Process is a method where Heartbeat processor checks whether a heartbeat metric is needed and injects it if so
func (h *Heartbeat) Process(metrics []telegraf.Metric) ([]telegraf.Metric, error) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.resetMap()

	minutes := findMetricMinutes(metrics)

	for minute, timeSeconds := range minutes {
		if h.heartbeatNeeded(minute) {
			newMetrics := h.addHeartbeat(metrics, minute, timeSeconds)
			metrics = newMetrics
		}
	}

	return metrics, nil
}

// Close clears the resources processor used, no-op in this case
func (h *Heartbeat) Close() {}

func findMetricMinutes(metrics []telegraf.Metric) map[int64]int64 {
	// holds a mapping between a minute and the "biggest" timestamp (in seconds) found for that minute
	minMap := make(map[int64]int64)

	for _, m := range metrics {
		min := getEpochMinute(m.Time())
		seconds := m.Time().Unix()

		if seconds > minMap[min] {
			minMap[min] = seconds
		}
	}

	return minMap
}

func (h *Heartbeat) addHeartbeat(metrics []telegraf.Metric, minute int64, timeSeconds int64) []telegraf.Metric {
	hb := buildHeartbeatMetric(time.Unix(timeSeconds, 0))

	metrics = append(metrics, hb)
	h.injectedMinutes[minute] = true

	return metrics
}

func buildHeartbeatMetric(timestamp time.Time) telegraf.Metric {
	// no need to inject any Sematext specific tags since MetricProcessors will be run afterwards and will take care
	// of such things
	hb := metric.New("heartbeat",
		make(map[string]string),
		map[string]interface{}{"alive": int64(1)},
		timestamp, telegraf.Gauge)

	return hb
}

func (h *Heartbeat) heartbeatNeeded(minute int64) bool {
	return !h.injectedMinutes[minute]
}

func (h *Heartbeat) resetMap() {
	day := getEpochDay(time.Now())

	if day > h.mapResetDay {
		h.injectedMinutes = make(map[int64]bool, oneDayMinutes)
		h.mapResetDay = day
	}
}

func getEpochDay(t time.Time) int64 {
	return t.Unix() / oneDaySeconds
}

func getEpochMinute(t time.Time) int64 {
	return t.Unix() / oneMinuteSeconds
}
