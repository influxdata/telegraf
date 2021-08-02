package processors

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestProcessMetric(t *testing.T) {
	now := time.Now()
	m := metric.New(
		"os",
		map[string]string{telegrafHostTag: "somehost", "os.disk": "sda1"},
		map[string]interface{}{"disk.used": float64(12.34), "disk.free": int64(55), "disk.size": uint64(777)},
		now)

	sentMetrics := make(map[string]*MetricMetainfo)

	mInfo, mKey := processMetric("aaa", m, getField(m, "disk.used"), sentMetrics)
	assert.NotNil(t, mInfo)

	sentMetrics[mKey] = mInfo
	// once it is in sentMetrics, it shouldn't be created again
	mInfo, mKey = processMetric("aaa", m, getField(m, "disk.used"), sentMetrics)
	assert.Nil(t, mInfo)

	sentMetrics[mKey] = nil
	m.RemoveTag(telegrafHostTag)
	mInfo, _ = processMetric("aaa", m, getField(m, "disk.used"), sentMetrics)
	assert.Nil(t, mInfo)
}

func TestBuildMetainfo(t *testing.T) {
	now := time.Now()
	m := metric.New(
		"os",
		map[string]string{telegrafHostTag: "somehost", "os.disk": "sda1"},
		map[string]interface{}{"disk.used": float64(12.34), "disk.free": int64(55), "disk.size": uint64(777)},
		now)

	mInfo := buildMetainfo("aaa", "host", m, getField(m, "disk.used"))

	assert.Equal(t, "aaa", mInfo.token)
	assert.Equal(t, "host", mInfo.host)
	assert.Equal(t, "os", mInfo.namespace)
	assert.Equal(t, "disk.used", mInfo.name)
	assert.Equal(t, Gauge, mInfo.semType)
	assert.Equal(t, Double, mInfo.numericType)
	assert.Equal(t, "os.disk.used", mInfo.label)
	assert.Equal(t, "", mInfo.description)

	m = metric.New(
		"os",
		map[string]string{telegrafHostTag: "somehost", "os.disk": "sda1"},
		map[string]interface{}{"disk.used": float64(12.34), "disk.free": int64(55), "disk.size": uint64(777)},
		now, telegraf.Counter)
	mInfo = buildMetainfo("aaa", "host", m, getField(m, "disk.used"))

	assert.Equal(t, Counter, mInfo.semType)

	m = metric.New(
		"os",
		map[string]string{telegrafHostTag: "somehost", "os.disk": "sda1"},
		map[string]interface{}{"disk.state": "active"},
		now, telegraf.Counter)
	mInfo = buildMetainfo("aaa", "host", m, getField(m, "disk.state"))
	assert.Nil(t, mInfo)
}

func getField(m telegraf.Metric, name string) *telegraf.Field {
	for _, f := range m.FieldList() {
		if f.Key == name {
			return f
		}
	}
	return nil
}

func TestGetSematextMetricType(t *testing.T) {
	assert.Equal(t, Gauge, getSematextMetricType(telegraf.Gauge))
	assert.Equal(t, Counter, getSematextMetricType(telegraf.Counter))
	assert.Equal(t, Gauge, getSematextMetricType(telegraf.Histogram))
	assert.Equal(t, Gauge, getSematextMetricType(telegraf.Summary))
	assert.Equal(t, Gauge, getSematextMetricType(telegraf.Untyped))
}

func TestGetSematextNumericType(t *testing.T) {
	assert.Equal(t, Double, getSematextNumericType(&telegraf.Field{
		Key:   "abc",
		Value: 1.23,
	}))
	assert.Equal(t, Long, getSematextNumericType(&telegraf.Field{
		Key:   "abc",
		Value: int64(11),
	}))
	assert.Equal(t, Long, getSematextNumericType(&telegraf.Field{
		Key:   "abc",
		Value: uint64(11),
	}))
	assert.Equal(t, Bool, getSematextNumericType(&telegraf.Field{
		Key:   "abc",
		Value: true,
	}))
	assert.Equal(t, UnsupportedNumericType, getSematextNumericType(&telegraf.Field{
		Key:   "abc",
		Value: "abc",
	}))
}

func TestBuildMetricKey(t *testing.T) {
	assert.Equal(t, "host-os.cpu.user", buildMetricKey("host", "os", "cpu.user"))
}
