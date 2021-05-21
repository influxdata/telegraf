package dynatrace

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/dynatrace-oss/dynatrace-metric-utils-go/metric/apiconstants"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestNilMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(`{"linesOk":10,"linesInvalid":0,"error":null}`)
	}))
	defer ts.Close()

	d := &Dynatrace{
		Timeout: config.Duration(time.Second * 5),
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
	require.NoError(t, err)
	require.Equal(t, apiconstants.GetDefaultOneAgentEndpoint(), d.URL)
	err = d.Connect()
	require.Equal(t, apiconstants.GetDefaultOneAgentEndpoint(), d.URL)
	require.NoError(t, err)
}

func TestMissingAPITokenMissingURL(t *testing.T) {
	d := &Dynatrace{}

	d.Log = testutil.Logger{}
	err := d.Init()
	require.NoError(t, err)
	require.Equal(t, apiconstants.GetDefaultOneAgentEndpoint(), d.URL)
	err = d.Connect()
	require.Equal(t, apiconstants.GetDefaultOneAgentEndpoint(), d.URL)
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
		expected := "mymeasurement.myfield,host=192.168.0.1,nix=nix gauge,3.14\nmymeasurement.value,host=192.168.0.1 count,3.14"
		if bodyString != expected {
			t.Errorf("Metric encoding failed. expected: %#v but got: %#v", expected, bodyString)
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

	m1 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1", "nix": "nix"},
		map[string]interface{}{"myfield": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	m2 := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
		telegraf.Counter,
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
		require.Regexp(t, regexp.MustCompile(`^mymeasurement\.myfield`), bodyString)
		require.Regexp(t, regexp.MustCompile(`a=test`), bodyString)
		require.Regexp(t, regexp.MustCompile(`b=test`), bodyString)
		require.Regexp(t, regexp.MustCompile(`c=test`), bodyString)
		require.Regexp(t, regexp.MustCompile(`gauge,3.14$`), bodyString)
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

	m1 := metric.New(
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
		expected := "mymeasurement.myfield gauge,3.14"
		if bodyString != expected {
			t.Errorf("Metric encoding failed. expected: %#v but got: %#v", expected, bodyString)
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

	m1 := metric.New(
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

		// expected := "mymeasurement.myfield,b_b=test,ccc=test,aaa=test gauge,3.14"
		// use regex because dimension order isn't guaranteed
		require.Regexp(t, regexp.MustCompile(`^mymeasurement\.myfield`), bodyString)
		require.Regexp(t, regexp.MustCompile(`aaa=test`), bodyString)
		require.Regexp(t, regexp.MustCompile(`b_b=test`), bodyString)
		require.Regexp(t, regexp.MustCompile(`ccc=test`), bodyString)
		require.Regexp(t, regexp.MustCompile(`gauge,3.14$`), bodyString)

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

	m1 := metric.New(
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
		// use regex because field order isn't guaranteed
		require.Contains(t, bodyString, "mymeasurement.yes gauge,1")
		require.Contains(t, bodyString, "mymeasurement.no gauge,0")
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

	m1 := metric.New(
		"mymeasurement",
		map[string]string{},
		map[string]interface{}{"yes": true, "no": false},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics := []telegraf.Metric{m1}

	err = d.Write(metrics)
	require.NoError(t, err)
}

func TestSendCounterMetricWithoutTags(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// check the encoded result
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			require.NoError(t, err)
		}
		bodyString := string(bodyBytes)
		expected := "mymeasurement.value gauge,32"
		if bodyString != expected {
			t.Errorf("Metric encoding failed. expected: %#v but got: %#v", expected, bodyString)
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

	m1 := metric.New(
		"mymeasurement",
		map[string]string{},
		map[string]interface{}{"value": 32},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics := []telegraf.Metric{m1}

	err = d.Write(metrics)
	require.NoError(t, err)
}
