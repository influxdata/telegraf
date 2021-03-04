package dynatrace

import (
	"encoding/json"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"io/ioutil"
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

	d := &Dynatrace{
		Timeout: internal.Duration{Duration: time.Second * 5},
	}

	d.URL = ts.URL
	d.APIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Init()
	require.NoError(t, err)

	err = d.Connect()
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

	d.URL = ts.URL
	d.APIToken = "123"
	d.Log = testutil.Logger{}

	err := d.Init()
	require.NoError(t, err)

	err = d.Connect()
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

	d.URL = ts.URL
	d.APIToken = "123"
	d.Log = testutil.Logger{}

	err := d.Init()
	require.NoError(t, err)
	err = d.Connect()
	require.NoError(t, err)
	err = d.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestMissingURL(t *testing.T) {
	d := &Dynatrace{}

	d.Log = testutil.Logger{}
	err := d.Init()
	require.Equal(t, oneAgentMetricsURL, d.URL)
	err = d.Connect()
	require.Equal(t, oneAgentMetricsURL, d.URL)
	require.NoError(t, err)
}

func TestMissingAPITokenMissingURL(t *testing.T) {
	d := &Dynatrace{}

	d.Log = testutil.Logger{}
	err := d.Init()
	require.Equal(t, oneAgentMetricsURL, d.URL)
	err = d.Connect()
	require.Equal(t, oneAgentMetricsURL, d.URL)
	require.NoError(t, err)
}

func TestMissingAPIToken(t *testing.T) {
	d := &Dynatrace{}

	d.URL = "test"
	d.Log = testutil.Logger{}
	err := d.Init()
	require.Error(t, err)
}

func TestSendMetric(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check the encoded result
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			require.NoError(t, err)
		}
		bodyString := string(bodyBytes)
		expected := "mymeasurement.myfield,host=\"192.168.0.1\",nix=\"nix\" 3.140000\nmymeasurement.value,host=\"192.168.0.1\" 3.140000\n"
		if bodyString != expected {
			t.Errorf("Metric encoding failed. expected: %s but got: %s", expected, bodyString)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"linesOk":10,"linesInvalid":0,"error":null}`)
	}))
	defer ts.Close()

	d := &Dynatrace{}

	d.URL = ts.URL
	d.APIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Init()
	require.NoError(t, err)
	err = d.Connect()
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

func TestSendSingleMetricWithUnorderedTags(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check the encoded result
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			require.NoError(t, err)
		}
		bodyString := string(bodyBytes)
		expected := "mymeasurement.myfield,a=\"test\",b=\"test\",c=\"test\" 3.140000\n"
		if bodyString != expected {
			t.Errorf("Metric encoding failed. expected: %s but got: %s", expected, bodyString)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"linesOk":1,"linesInvalid":0,"error":null}`)
	}))
	defer ts.Close()

	d := &Dynatrace{}

	d.URL = ts.URL
	d.APIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Init()
	require.NoError(t, err)
	err = d.Connect()
	require.NoError(t, err)

	// Init metrics

	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{"a": "test", "c": "test", "b": "test"},
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
		// check the encoded result
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			require.NoError(t, err)
		}
		bodyString := string(bodyBytes)
		expected := "mymeasurement.myfield 3.140000\n"
		if bodyString != expected {
			t.Errorf("Metric encoding failed. expected: %s but got: %s", expected, bodyString)
		}
		json.NewEncoder(w).Encode(`{"linesOk":1,"linesInvalid":0,"error":null}`)
	}))
	defer ts.Close()

	d := &Dynatrace{}

	d.URL = ts.URL
	d.APIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Init()
	require.NoError(t, err)
	err = d.Connect()
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

func TestSendMetricWithUpperCaseTagKeys(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// check the encoded result
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			require.NoError(t, err)
		}
		bodyString := string(bodyBytes)
		expected := "mymeasurement.myfield,aaa=\"test\",b_b=\"test\",ccc=\"test\" 3.140000\n"
		if bodyString != expected {
			t.Errorf("Metric encoding failed. expected: %s but got: %s", expected, bodyString)
		}
		json.NewEncoder(w).Encode(`{"linesOk":1,"linesInvalid":0,"error":null}`)
	}))
	defer ts.Close()

	d := &Dynatrace{}

	d.URL = ts.URL
	d.APIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Init()
	require.NoError(t, err)
	err = d.Connect()
	require.NoError(t, err)

	// Init metrics

	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{"AAA": "test", "CcC": "test", "B B": "test"},
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
		// check the encoded result
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			require.NoError(t, err)
		}
		bodyString := string(bodyBytes)
		expected := "mymeasurement.myfield 1\n"
		if bodyString != expected {
			t.Errorf("Metric encoding failed. expected: %s but got: %s", expected, bodyString)
		}
		json.NewEncoder(w).Encode(`{"linesOk":1,"linesInvalid":0,"error":null}`)
	}))
	defer ts.Close()

	d := &Dynatrace{}

	d.URL = ts.URL
	d.APIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Init()
	require.NoError(t, err)
	err = d.Connect()
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
