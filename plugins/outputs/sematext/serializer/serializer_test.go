package serializer

import (
	"fmt"
	"github.com/influxdata/telegraf/testutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
)

func TestLinePerMetricSerializerWrite(t *testing.T) {
	serializer := NewLinePerMetricSerializer(testutil.Logger{})

	now := time.Now()

	m := metric.New(
		"os",
		map[string]string{"os.host": "hostname"},
		map[string]interface{}{"disk.size": uint64(777)},
		now)

	metrics := []telegraf.Metric{m}

	assert.Equal(t,
		fmt.Sprintf("os,os.host=hostname disk.size=777i %d\n", now.UnixNano()),
		string(serializer.Write(metrics)))

	m = metric.New(
		"system",
		map[string]string{"os.host": "hostname", "token": "token"},
		map[string]interface{}{"uptime_format": "18 days, 22:37"},
		now)

	metrics = []telegraf.Metric{m}
	assert.Equal(t,
		// strings are temporarily filtered out (until Sematext backend gets support)
		// fmt.Sprintf("system,os.host=hostname,token=token uptime_format=\"18 days, 22:37\" %d\n", now.UnixNano()),
		"",
		string(serializer.Write(metrics)))
}

func TestLinePerMetricSerializerWriteNoTags(t *testing.T) {
	serializer := NewLinePerMetricSerializer(testutil.Logger{})

	now := time.Now()

	m := metric.New(
		"os",
		map[string]string{},
		map[string]interface{}{"disk.size": uint64(777)},
		now)

	metrics := []telegraf.Metric{m}

	assert.Equal(t,
		fmt.Sprintf("os disk.size=777i %d\n", now.UnixNano()),
		string(serializer.Write(metrics)))
}

func TestLinePerMetricSerializerWriteNoMetrics(t *testing.T) {
	serializer := NewLinePerMetricSerializer(testutil.Logger{})

	now := time.Now()

	m := metric.New(
		"os",
		map[string]string{"os.host": "hostname"},
		map[string]interface{}{},
		now)

	metrics := []telegraf.Metric{m}

	assert.Equal(t, "", string(serializer.Write(metrics)))
}

func TestLinePerMetricSerializerWriteMultipleTagsAndMetrics(t *testing.T) {
	serializer := NewLinePerMetricSerializer(testutil.Logger{})

	now := time.Now()

	m := metric.New(
		"os",
		map[string]string{"os.host": "hostname", "os.disk": "sda1"},
		map[string]interface{}{"disk.used": float64(12.34), "disk.free": int64(55), "disk.size": uint64(777)},
		now)

	metrics := []telegraf.Metric{m}

	assert.Equal(t,
		fmt.Sprintf("os,os.disk=sda1,os.host=hostname disk.free=55i,disk.size=777i,disk.used=12.34 %d\n", now.UnixNano()),
		string(serializer.Write(metrics)))
}

func TestCompactMetricSerializerWrite(t *testing.T) {
	serializer := NewCompactMetricSerializer(testutil.Logger{})

	now := time.Now()

	m := metric.New(
		"os",
		map[string]string{"os.host": "hostname"},
		map[string]interface{}{"disk.size": uint64(777)},
		now)

	metrics := []telegraf.Metric{m}

	assert.Equal(t,
		fmt.Sprintf("os,os.host=hostname disk.size=777i %d\n", now.UnixNano()),
		string(serializer.Write(metrics)))

	m = metric.New(
		"system",
		map[string]string{"os.host": "hostname", "token": "token"},
		map[string]interface{}{"uptime_format": "18 days, 22:37"},
		now)

	metrics = []telegraf.Metric{m}
	assert.Equal(t,
		// strings are temporarily filtered out (until Sematext backend gets support)
		// fmt.Sprintf("system,os.host=hostname,token=token uptime_format=\"18 days, 22:37\" %d\n", now.UnixNano()),
		"",
		string(serializer.Write(metrics)))
}

func TestCompactMetricSerializerWriteNoTags(t *testing.T) {
	serializer := NewCompactMetricSerializer(testutil.Logger{})

	now := time.Now()

	m := metric.New(
		"os",
		map[string]string{},
		map[string]interface{}{"disk.size": uint64(777)},
		now)

	metrics := []telegraf.Metric{m}

	assert.Equal(t,
		fmt.Sprintf("os disk.size=777i %d\n", now.UnixNano()),
		string(serializer.Write(metrics)))
}

func TestCompactMetricSerializerWriteNoMetrics(t *testing.T) {
	serializer := NewCompactMetricSerializer(testutil.Logger{})

	now := time.Now()

	m := metric.New(
		"os",
		map[string]string{"os.host": "hostname"},
		map[string]interface{}{},
		now)

	metrics := []telegraf.Metric{m}

	assert.Equal(t, "", string(serializer.Write(metrics)))
}

func TestCompactMetricSerializerWriteMultipleTagsAndMetrics(t *testing.T) {
	serializer := NewCompactMetricSerializer(testutil.Logger{})

	now := time.Now()

	m := metric.New(
		"os",
		map[string]string{"os.host": "hostname", "os.disk": "sda1"},
		map[string]interface{}{"disk.used": float64(12.34), "disk.free": int64(55), "disk.size": uint64(777)},
		now)

	metrics := []telegraf.Metric{m}

	assert.Equal(t,
		fmt.Sprintf("os,os.disk=sda1,os.host=hostname disk.free=55i,disk.size=777i,disk.used=12.34 %d\n", now.UnixNano()),
		string(serializer.Write(metrics)))
}

func TestCompactMetricSerializerWriteMultipleMetricsSingleLine(t *testing.T) {
	serializer := NewCompactMetricSerializer(testutil.Logger{})

	now := time.Now()

	m1 := metric.New(
		"os",
		map[string]string{"os.host": "hostname", "os.disk": "sda1"},
		map[string]interface{}{"disk.used": float64(12.34)},
		now)

	m2 := metric.New(
		"os",
		map[string]string{"os.host": "hostname", "os.disk": "sda1"},
		map[string]interface{}{"disk.free": int64(55)},
		now)

	m3 := metric.New(
		"os",
		map[string]string{"os.host": "hostname", "os.disk": "sda1"},
		map[string]interface{}{"disk.size": uint64(777)},
		now)

	metrics := []telegraf.Metric{m1, m2, m3}

	assert.Equal(t,
		fmt.Sprintf("os,os.disk=sda1,os.host=hostname disk.free=55i,disk.size=777i,disk.used=12.34 %d\n", now.UnixNano()),
		string(serializer.Write(metrics)))
}

func TestCompactMetricSerializerWriteMultipleMetricsMultipleLines(t *testing.T) {
	serializer := NewCompactMetricSerializer(testutil.Logger{})

	now := time.Now()

	m1 := metric.New(
		"os",
		map[string]string{"os.host": "hostname", "os.disk": "sda1"},
		map[string]interface{}{"disk.used": float64(12.34)},
		now)

	m2 := metric.New(
		"os",
		map[string]string{"os.host": "hostname", "os.disk": "sda1"},
		map[string]interface{}{"disk.free": int64(55)},
		now)

	m3 := metric.New(
		"somethingelse",
		map[string]string{"os.host": "hostname", "os.disk": "sda1"},
		map[string]interface{}{"disk.size": uint64(777)},
		now)

	metrics := []telegraf.Metric{m1, m2, m3}

	assert.Equal(t,
		fmt.Sprintf("os,os.disk=sda1,os.host=hostname disk.free=55i,disk.used=12.34 %d\n"+
			"somethingelse,os.disk=sda1,os.host=hostname disk.size=777i %d\n", now.UnixNano(), now.UnixNano()),
		string(serializer.Write(metrics)))
}

func TestCompactMetricSerializerWriteMultipleMetricsDifferentTimestamp(t *testing.T) {
	serializer := NewCompactMetricSerializer(testutil.Logger{})

	now := time.Now()

	m1 := metric.New(
		"os",
		map[string]string{"os.host": "hostname", "os.disk": "sda1"},
		map[string]interface{}{"disk.used": float64(12.34)},
		now)

	now2 := time.Now().AddDate(0, 0, 1)

	m2 := metric.New(
		"os",
		map[string]string{"os.host": "hostname", "os.disk": "sda1"},
		map[string]interface{}{"disk.free": int64(55)},
		now2)

	now3 := time.Now().AddDate(0, 0, 2)

	m3 := metric.New(
		"os",
		map[string]string{"os.host": "hostname", "os.disk": "sda1"},
		map[string]interface{}{"disk.size": uint64(777)},
		now3)

	metrics := []telegraf.Metric{m1, m2, m3}

	assert.Equal(t,
		fmt.Sprintf("os,os.disk=sda1,os.host=hostname disk.used=12.34 %d\n"+
			"os,os.disk=sda1,os.host=hostname disk.free=55i %d\n"+
			"os,os.disk=sda1,os.host=hostname disk.size=777i %d\n",
			now.UnixNano(), now2.UnixNano(), now3.UnixNano()),
		string(serializer.Write(metrics)))
}
