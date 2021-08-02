package sematext

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestCheckResponseStatus(t *testing.T) {
	res := &http.Response{
		StatusCode: 200,
	}
	success, badRequest := checkResponseStatus(res)
	assert.Equal(t, true, success)
	assert.Equal(t, false, badRequest)

	res.StatusCode = 301
	success, badRequest = checkResponseStatus(res)
	assert.Equal(t, false, success)
	assert.Equal(t, false, badRequest)

	res.StatusCode = 404
	success, badRequest = checkResponseStatus(res)
	assert.Equal(t, false, success)
	assert.Equal(t, true, badRequest)

	res.StatusCode = 500
	success, badRequest = checkResponseStatus(res)
	assert.Equal(t, false, success)
	assert.Equal(t, false, badRequest)
}

func TestHandleResponse(t *testing.T) {
	sem := &Sematext{
		Log: testutil.Logger{},
	}
	res := &http.Response{
		StatusCode: 200,
	}
	assert.Nil(t, sem.handleResponse(res))

	res.StatusCode = 301
	assert.NotNil(t, sem.handleResponse(res))

	res.StatusCode = 404
	assert.Nil(t, sem.handleResponse(res))

	res.StatusCode = 500
	assert.NotNil(t, sem.handleResponse(res))
}

func TestMetricAlreadyProcessed(t *testing.T) {
	now := time.Now()
	m := metric.New(
		"os",
		map[string]string{"os.host": "somehost", "os.disk": "sda1"},
		map[string]interface{}{"disk.used": float64(12.34)},
		now)

	assert.False(t, metricAlreadyProcessed(m))

	metrics := make([]telegraf.Metric, 0)
	metrics = append(metrics, m)

	markMetricsProcessed(metrics)

	assert.True(t, metricAlreadyProcessed(m))
}

func TestMetricsAlreadyProcessed(t *testing.T) {
	now := time.Now()
	m1 := metric.New(
		"os",
		map[string]string{"os.host": "somehost", "os.disk": "sda1"},
		map[string]interface{}{"disk.used": float64(12.34)},
		now)

	m2 := metric.New(
		"os",
		map[string]string{"os.host": "somehost", "os.disk": "sda2"},
		map[string]interface{}{"disk.used": float64(23.45)},
		now)

	metrics := make([]telegraf.Metric, 0)
	metrics = append(metrics, m1, m2)

	assert.False(t, metricsAlreadyProcessed(metrics))

	markMetricsProcessed(metrics)

	assert.True(t, metricsAlreadyProcessed(metrics))
}
