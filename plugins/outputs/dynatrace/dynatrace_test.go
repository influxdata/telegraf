package dynatrace

import (
	"encoding/json"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNilMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"linesOk":10,"linesInvalid":0,"error":null}`)
	}))
	defer ts.Close()

	d := &Dynatrace{}
	d.EnvironmentURL = ts.URL
	d.EnvironmentAPIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Connect()
	require.NoError(t, err)

	err = d.Write(nil)
	require.NoError(t, err)
}

func TestEmptyMetricsSlice(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"linesOk":10,"linesInvalid":0,"error":null}`)
	}))
	defer ts.Close()

	d := &Dynatrace{}
	d.EnvironmentURL = ts.URL
	d.EnvironmentAPIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Connect()
	require.NoError(t, err)
	empty := []telegraf.Metric{}
	err = d.Write(empty)
	require.NoError(t, err)
}

func TestMockURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"linesOk":10,"linesInvalid":0,"error":null}`)
	}))
	defer ts.Close()

	d := &Dynatrace{}
	d.EnvironmentURL = ts.URL
	d.EnvironmentAPIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Connect()
	require.NoError(t, err)

	err = d.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestMissingURL(t *testing.T) {
	d := &Dynatrace{}
	d.EnvironmentAPIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Connect()
	require.Error(t, err)
}

func TestMissingAPIToken(t *testing.T) {
	d := &Dynatrace{}
	d.EnvironmentURL = "test"
	d.Log = testutil.Logger{}
	err := d.Connect()
	require.Error(t, err)
}

func TestSendMetric(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"linesOk":10,"linesInvalid":0,"error":null}`)
	}))
	defer ts.Close()

	d := &Dynatrace{}
	d.EnvironmentURL = ts.URL
	d.EnvironmentAPIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Connect()
	require.NoError(t, err)

	// Init metrics

	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1", "nix": "nix"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	m2, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics := []telegraf.Metric{m1, m2}

	err = d.Write(metrics)
	require.NoError(t, err)
}

func TestSendSingleMetric(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"linesOk":10,"linesInvalid":0,"error":null}`)
	}))
	defer ts.Close()

	d := &Dynatrace{}
	d.EnvironmentURL = ts.URL
	d.EnvironmentAPIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Connect()
	require.NoError(t, err)

	// Init metrics

	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1", "nix": "nix"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics := []telegraf.Metric{m1}

	err = d.Write(metrics)
	require.NoError(t, err)
}

func TestSendMetricWithoutTags(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"linesOk":10,"linesInvalid":0,"error":null}`)
	}))
	defer ts.Close()

	d := &Dynatrace{}
	d.EnvironmentURL = ts.URL
	d.EnvironmentAPIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Connect()
	require.NoError(t, err)

	// Init metrics

	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics := []telegraf.Metric{m1}

	err = d.Write(metrics)
	require.NoError(t, err)
}

func TestSendBooleanMetricWithoutTags(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"linesOk":10,"linesInvalid":0,"error":null}`)
	}))
	defer ts.Close()

	d := &Dynatrace{}
	d.EnvironmentURL = ts.URL
	d.EnvironmentAPIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Connect()
	require.NoError(t, err)

	// Init metrics

	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{},
		map[string]interface{}{"myfield": bool(true)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics := []telegraf.Metric{m1}

	err = d.Write(metrics)
	require.NoError(t, err)
}
