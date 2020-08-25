package cloudfoundry

import (
	"testing"
	"time"

	"code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestCastTimerMetric(t *testing.T) {
	ts := time.Now()
	env := &loggregator_v2.Envelope{
		SourceId:   "source",
		InstanceId: "instance",
		Timestamp:  ts.UnixNano(),
		Message: &loggregator_v2.Envelope_Timer{
			Timer: &loggregator_v2.Timer{
				Name:  "http",
				Start: 1 * int64(time.Second),
				Stop:  8 * int64(time.Second),
			},
		},
		Tags: map[string]string{
			"uri":         "http://example/uri",
			"app_name":    "app1",
			"status_code": "200",
		},
	}

	want := testutil.MustMetric(
		CloudfoundryMeasurement,
		map[string]string{
			"app_name":    "app1",
			"source_id":   "source",
			"instance_id": "instance",
		},
		map[string]interface{}{
			"uri":           "http://example/uri",
			"status_code":   int64(200),
			"http_start":    int64(1000000000),
			"http_stop":     int64(8000000000),
			"http_duration": int64(7000000000),
		},
		ts.UTC(),
		telegraf.Untyped,
	)

	m, err := NewMetric(env)
	require.NoError(t, err)

	testutil.RequireMetricEqual(t, want, m)
}

func TestCastLogMetric(t *testing.T) {
	ts := time.Now()
	env := &loggregator_v2.Envelope{
		SourceId:   "source",
		InstanceId: "instance",
		Timestamp:  ts.UnixNano(),
		Message: &loggregator_v2.Envelope_Log{
			Log: &loggregator_v2.Log{
				Type:    loggregator_v2.Log_OUT,
				Payload: []byte("stdout log msg"),
			},
		},
		Tags: map[string]string{
			"source_type":       "APP/WEB/0",
			"organization_name": "org1",
			"space_name":        "space1",
			"app_name":          "app1",
		},
	}

	want := testutil.MustMetric(
		SyslogMeasurement,
		map[string]string{
			"source_id":         "source",
			"source_type":       "APP/WEB/0",
			"instance_id":       "instance",
			"organization_name": "org1",
			"space_name":        "space1",
			"app_name":          "app1",
			"hostname":          "org1.space1.app1",
			"appname":           "app1",
			"severity":          "info",
			"facility":          "user",
		},
		map[string]interface{}{
			"message":       "stdout log msg",
			"timestamp":     ts.UnixNano(),
			"facility_code": int64(1),
			"severity_code": int64(6),
			"procid":        "APP/WEB/0",
			"version":       int64(1),
		},
		ts.UTC(),
		telegraf.Untyped,
	)

	m, err := NewMetric(env)
	require.NoError(t, err)

	testutil.RequireMetricEqual(t, want, m)
}

func TestCastCounterMetric(t *testing.T) {
	ts := time.Now()
	env := &loggregator_v2.Envelope{
		SourceId:   "source",
		InstanceId: "instance",
		Timestamp:  ts.UnixNano(),
		Message: &loggregator_v2.Envelope_Counter{
			Counter: &loggregator_v2.Counter{
				Name:  "counter",
				Total: 100,
				Delta: 1,
			},
		},
		Tags: map[string]string{
			"organization_name": "org1",
			"space_name":        "space1",
			"app_name":          "app1",
		},
	}

	want := testutil.MustMetric(
		CloudfoundryMeasurement,
		map[string]string{
			"source_id":         "source",
			"instance_id":       "instance",
			"organization_name": "org1",
			"space_name":        "space1",
			"app_name":          "app1",
		},
		map[string]interface{}{
			"counter_total": uint64(100),
			"counter_delta": uint64(1),
		},
		ts.UTC(),
		telegraf.Counter,
	)

	m, err := NewMetric(env)
	require.NoError(t, err)

	testutil.RequireMetricEqual(t, want, m)
}

func TestCastGaugeMetric(t *testing.T) {
	ts := time.Now()
	env := &loggregator_v2.Envelope{
		SourceId:   "source",
		InstanceId: "instance",
		Timestamp:  ts.UnixNano(),
		Message: &loggregator_v2.Envelope_Gauge{
			Gauge: &loggregator_v2.Gauge{
				Metrics: map[string]*loggregator_v2.GaugeValue{
					"cpu": {
						Unit:  "ns",
						Value: float64(1),
					},
					"mem": {
						Unit:  "mb",
						Value: float64(1024),
					},
				},
			},
		},
		Tags: map[string]string{
			"organization_name": "org1",
			"space_name":        "space1",
			"app_name":          "app1",
		},
	}

	want := testutil.MustMetric(
		CloudfoundryMeasurement,
		map[string]string{
			"source_id":         "source",
			"instance_id":       "instance",
			"organization_name": "org1",
			"space_name":        "space1",
			"app_name":          "app1",
		},
		map[string]interface{}{
			"cpu": float64(1),
			"mem": float64(1024),
		},
		ts.UTC(),
		telegraf.Gauge,
	)

	m, err := NewMetric(env)
	require.NoError(t, err)

	testutil.RequireMetricEqual(t, want, m)
}

func TestCastEventMetric(t *testing.T) {
	ts := time.Now()
	env := &loggregator_v2.Envelope{
		SourceId:   "source",
		InstanceId: "instance",
		Timestamp:  ts.UnixNano(),
		Message: &loggregator_v2.Envelope_Event{
			Event: &loggregator_v2.Event{
				Title: "thing_occurred",
				Body:  "a thing has occurred",
			},
		},
		Tags: map[string]string{
			"organization_name": "org1",
			"space_name":        "space1",
			"app_name":          "app1",
		},
	}

	want := testutil.MustMetric(
		CloudfoundryMeasurement,
		map[string]string{
			"source_id":         "source",
			"instance_id":       "instance",
			"organization_name": "org1",
			"space_name":        "space1",
			"app_name":          "app1",
		},
		map[string]interface{}{
			"title": "thing_occurred",
			"body":  "a thing has occurred",
		},
		ts.UTC(),
		telegraf.Untyped,
	)

	m, err := NewMetric(env)
	require.NoError(t, err)

	testutil.RequireMetricEqual(t, want, m)
}
