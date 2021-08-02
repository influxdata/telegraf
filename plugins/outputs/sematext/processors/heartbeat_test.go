package processors

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
)

func TestBuildHeartbeatMetric(t *testing.T) {
	now := time.Now()
	m := buildHeartbeatMetric(now)

	assert.Equal(t, "heartbeat", m.Name())
	assert.Equal(t, 1, len(m.Fields()))
	val, set := m.GetField("alive")
	assert.Equal(t, true, set)
	assert.Equal(t, int64(1), val)
	assert.Equal(t, 0, len(m.Tags()))
}

func TestHeartbeatNeeded(t *testing.T) {
	minute := int64(11)
	bp := NewHeartbeat()
	h := bp.(*Heartbeat)
	assert.Equal(t, true, h.heartbeatNeeded(minute))

	h.injectedMinutes[minute] = true
	assert.Equal(t, false, h.heartbeatNeeded(minute))

	minute++
	assert.Equal(t, true, h.heartbeatNeeded(minute))
}

func TestAddHeartbeat(t *testing.T) {
	bp := NewHeartbeat()
	h := bp.(*Heartbeat)
	now := time.Now()
	currentMinute := getEpochMinute(now)

	assert.Equal(t, false, h.injectedMinutes[currentMinute])
	metrics := make([]telegraf.Metric, 0, 1)
	var err error
	metrics = h.addHeartbeat(metrics, currentMinute, now.Unix())

	assert.Nil(t, err)
	assert.Equal(t, true, h.injectedMinutes[currentMinute])
	assert.Equal(t, 1, len(metrics))
	assert.Equal(t, now.Unix(), metrics[0].Time().Unix())
}

func TestProcess(t *testing.T) {
	bp := NewHeartbeat()
	h := bp.(*Heartbeat)
	metrics := make([]telegraf.Metric, 0, 2)

	var err error
	metrics, err = h.Process(metrics)

	// no metrics, so no heartbeat metric should be injected
	assert.Nil(t, err)
	assert.Equal(t, 0, len(metrics))

	now := time.Now()
	currentMinute := getEpochMinute(now)
	cpuMetric := metric.New("os",
		make(map[string]string),
		map[string]interface{}{"cpu.user": 99.9},
		now, telegraf.Gauge)

	assert.Nil(t, err)

	metrics = append(metrics, cpuMetric)

	metrics, err = h.Process(metrics)
	assert.Nil(t, err)
	assert.Equal(t, true, h.injectedMinutes[currentMinute])
	// cpu.user and a heartbeat metric:
	assert.Equal(t, 2, len(metrics))
}

func TestFindMetricMinutes(t *testing.T) {
	metrics := make([]telegraf.Metric, 0, 1)

	now := time.Now()
	currentMinute := getEpochMinute(now)
	cpuMetric := metric.New("os",
		make(map[string]string),
		map[string]interface{}{"cpu.user": 99.9},
		now, telegraf.Gauge)

	metrics = append(metrics, cpuMetric)

	minMap := findMetricMinutes(metrics)
	assert.Equal(t, now.Unix(), minMap[currentMinute])
}

func TestResetMap(t *testing.T) {
	bp := NewHeartbeat()
	h := bp.(*Heartbeat)
	h.injectedMinutes[123] = true
	h.mapResetDay = 123

	h.resetMap()

	assert.Equal(t, 0, len(h.injectedMinutes))
	assert.NotEqual(t, 123, h.mapResetDay)
}
