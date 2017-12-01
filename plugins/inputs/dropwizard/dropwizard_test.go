package dropwizard_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/plugins/inputs/dropwizard"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestDecodeGaugeIntValueJSON(t *testing.T) {
	plugin := &dropwizard.Dropwizard{}
	metrics, err := plugin.DecodeJSONMetrics(strings.NewReader(gaugeWithIntValueJSON))
	if err != nil {
		t.Fatalf("error parsing json %v", err)
	}
	if len(metrics.Gauges) != 1 {
		t.Errorf("No. of gauges was incorrect, got: %d, want: 1.", len(metrics.Gauges))
	}

	memoryUsedGauge := metrics.Gauges["jvm.memory.total.used"]
	
	if memoryUsedGauge.Value.Type != dropwizard.IntType {
		t.Error("Memory gauge value was not an int.")
	}

	if memoryUsedGauge.Value.IntValue != 77233584 {
		t.Errorf("Memory gauge value was incorrect, got: %d, want: 77233584.", memoryUsedGauge.Value.IntValue)
	}
}

func TestDecodeGaugeFloatValueJSON(t *testing.T) {
	plugin := &dropwizard.Dropwizard{}
	metrics, err := plugin.DecodeJSONMetrics(strings.NewReader(gaugeWithFloatValueJSON))
	if err != nil {
		t.Fatalf("error parsing json %v", err)
	}
	if len(metrics.Gauges) != 1 {
		t.Errorf("No. of gauges was incorrect, got: %d, want: 1.", len(metrics.Gauges))
	}

	memoryUsedGauge := metrics.Gauges["gauge.float.example"]
	
	if memoryUsedGauge.Value.Type != dropwizard.FloatType {
		t.Error("Memory gauge value was not an float.")
	}

	if memoryUsedGauge.Value.FloatValue != 50.099579601 {
		t.Errorf("Memory gauge value was incorrect, got: %f, want: 50.099579601.", memoryUsedGauge.Value.FloatValue)
	}
}

func TestDecodeGaugeStringValueJSON(t *testing.T) {
	plugin := &dropwizard.Dropwizard{}
	metrics, err := plugin.DecodeJSONMetrics(strings.NewReader(gaugeWithStringValueJSON))
	if err != nil {
		t.Fatalf("error parsing json %v", err)
	}
	if len(metrics.Gauges) != 1 {
		t.Errorf("No. of gauges was incorrect, got: %d, want: 1.", len(metrics.Gauges))
	}

	memoryUsedGauge := metrics.Gauges["io.dropwizard.jetty.MutableServletContextHandler.percent-4xx-15m"]
	
	if memoryUsedGauge.Value.Type != dropwizard.StringType {
		t.Error("Gauge value was an int. Expected string.")
	}

	if memoryUsedGauge.Value.StringValue != "NaN" {
		t.Errorf("Gauge value was incorrect, got: %s, want: NaN.", memoryUsedGauge.Value.StringValue)
	}
}

func TestDecodeFloatsInTimersJSON(t *testing.T) {
	plugin := &dropwizard.Dropwizard{}
	metrics, err := plugin.DecodeJSONMetrics(strings.NewReader(oneMetricPerTypeJSON))
	if err != nil {
		t.Fatalf("error parsing json %v", err)
	}
	if len(metrics.Timers) != 1 {
		t.Errorf("No. of timers was incorrect, got: %d, want: 1.", len(metrics.Gauges))
	}

	connectionsTimer := metrics.Timers["org.eclipse.jetty.server.HttpConnectionFactory.8081.connections"]
	
	if connectionsTimer.Mean != 50.13867439111584 {
		t.Errorf("Connections timer mean was incorrect, got: %.2f, want: 50.13867439111584.", connectionsTimer.Mean)
	}
}

func TestBasic(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, _ = w.Write([]byte(oneMetricPerTypeJSON))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	plugin := &dropwizard.Dropwizard{
		URLs: []string{fakeServer.URL + "/endpoint"},
	}

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))

	require.Len(t, acc.Metrics, 5)

	fields := map[string]interface{}{
		"value": int64(77233584),
	}
	acc.AssertContainsFields(t, "jvm.memory.total.used", fields)
	
	fields = map[string]interface{}{
		"count": int64(0),
	}
	acc.AssertContainsFields(t, "io.dropwizard.jetty.MutableServletContextHandler.active-dispatches", fields)

	fields = map[string]interface{}{
		"count": int64(3),
		"max" : int64(0),
		"mean" : float64(0.0),
		"min" : int64(0),
		"p50" : float64(0.0),
		"p75" : float64(0.0),
		"p95" : float64(0.0),
		"p98" : float64(0.0),
		"p99" : float64(0.0),
		"p999" : float64(0.0),
		"stddev" : float64(0.0),
	}
	acc.AssertContainsFields(t, "histogram.example", fields)

	fields = map[string]interface{}{
		"count": int64(0),
		"m15_rate" : float64(0.0),
		"m1_rate" : float64(0.0),
		"m5_rate" : float64(0.0),
		"mean_rate" : float64(0.0),
	}
	acc.AssertContainsFields(t, "ch.qos.logback.core.Appender.error", fields)

	fields = map[string]interface{}{
		"count": int64(2),
		"max" : float64(82.058711464),
		"mean" : float64(50.13867439111584),
		"min" : float64(50.099579601),
		"p50" : float64(50.099579601),
		"p75" : float64(50.099579601),
		"p95" : float64(50.099579601),
		"p98" : float64(50.099579601),
		"p99" : float64(50.099579601),
		"p999" : float64(82.058711464),
		"stddev" : float64(1.1170976456220436),
		"m15_rate" : float64(0.0017641182197863407),
		"m1_rate" : float64(0.01354432563862594),
		"m5_rate" : float64(0.003922747115045747),
		"mean_rate" : float64(0.003482175245668154),
	}
	acc.AssertContainsFields(t, "org.eclipse.jetty.server.HttpConnectionFactory.8081.connections", fields)
}

func TestSkippingIdleMetrics(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, _ = w.Write([]byte(oneMetricPerTypeJSON))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	plugin := &dropwizard.Dropwizard{
		URLs: []string{fakeServer.URL + "/endpoint"},
		SkipIdleMetrics: true,
	}

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))

	require.Len(t, acc.Metrics, 5)

	acc.ClearMetrics()

	require.NoError(t, acc.GatherError(plugin.Gather))

	// only the gauge should come through
	require.Len(t, acc.Metrics, 1)
}

func TestFloatFormatting(t *testing.T) {
	plugin := &dropwizard.Dropwizard{
		FloatFieldFormat: "%.2f",
	}
	f := plugin.FormatFloat(82.058711464)

	// it will round up
	if f != 82.06 {
		t.Errorf("Formatted float was incorrect, got: %.9f, want: 82.06.", f)
	}
}	

const oneMetricPerTypeJSON = `
{
	"version" : "3.1.3",
	"gauges" : {
		"jvm.memory.total.used" : {
			"value" : 77233584
		}
	},
	"counters" : {
		"io.dropwizard.jetty.MutableServletContextHandler.active-dispatches" : {
			"count" : 0
		}
	},
	"histograms" : {
		"histogram.example" : {
			"count" : 3,
			"max" : 0,
			"mean" : 0.0,
			"min" : 0,
			"p50" : 0.0,
			"p75" : 0.0,
			"p95" : 0.0,
			"p98" : 0.0,
			"p99" : 0.0,
			"p999" : 0.0,
			"stddev" : 0.0
		}
	},
	"meters" : {
		"ch.qos.logback.core.Appender.error" : {
			"count" : 0,
			"m15_rate" : 0.0,
			"m1_rate" : 0.0,
			"m5_rate" : 0.0,
			"mean_rate" : 0.0,
			"units" : "events/second"
		}
	},
	"timers" : {
		"org.eclipse.jetty.server.HttpConnectionFactory.8081.connections" : {
			"count" : 2,
			"max" : 82.058711464,
			"mean" : 50.13867439111584,
			"min" : 50.099579601,
			"p50" : 50.099579601,
			"p75" : 50.099579601,
			"p95" : 50.099579601,
			"p98" : 50.099579601,
			"p99" : 50.099579601,
			"p999" : 82.058711464,
			"stddev" : 1.1170976456220436,
			"m15_rate" : 0.0017641182197863407,
			"m1_rate" : 0.01354432563862594,
			"m5_rate" : 0.003922747115045747,
			"mean_rate" : 0.003482175245668154,
			"duration_units" : "seconds",
			"rate_units" : "calls/second"
		}
	}
}
`

const gaugeWithStringValueJSON = `
{
	"version" : "3.1.3",
	"gauges" : {
		"io.dropwizard.jetty.MutableServletContextHandler.percent-4xx-15m" : {
			"value" : "NaN"
		}
	},
	"counters" : { },
	"histograms" : { },
	"meters" : { },
	"timers" : { }
}
`

const gaugeWithIntValueJSON = `
{
	"version" : "3.1.3",
	"gauges" : {
		"jvm.memory.total.used" : {
			"value" : 77233584
		}
	},
	"counters" : { },
	"histograms" : { },
	"meters" : { },
	"timers" : { }
}
`

const gaugeWithFloatValueJSON = `
{
	"version" : "3.1.3",
	"gauges" : {
		"gauge.float.example" : {
			"value" : 50.099579601
		}
	},
	"counters" : { },
	"histograms" : { },
	"meters" : { },
	"timers" : { }
}
`