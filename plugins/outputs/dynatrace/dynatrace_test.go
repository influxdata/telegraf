package dynatrace

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/dynatrace-oss/dynatrace-metric-utils-go/metric/apiconstants"
	"github.com/dynatrace-oss/dynatrace-metric-utils-go/metric/dimensions"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestNilMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(`{"linesOk":10,"linesInvalid":0,"error":null}`)
		require.NoError(t, err)
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
		err := json.NewEncoder(w).Encode(`{"linesOk":10,"linesInvalid":0,"error":null}`)
		require.NoError(t, err)
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
		err := json.NewEncoder(w).Encode(`{"linesOk":10,"linesInvalid":0,"error":null}`)
		require.NoError(t, err)
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

func TestSendMetrics(t *testing.T) {
	expected := []string{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check the encoded result
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		bodyString := string(bodyBytes)

		lines := strings.Split(bodyString, "\n")

		sort.Strings(lines)
		sort.Strings(expected)

		expectedString := strings.Join(expected, "\n")
		foundString := strings.Join(lines, "\n")
		if foundString != expectedString {
			t.Errorf("Metric encoding failed. expected: %#v but got: %#v", expectedString, foundString)
		}
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(fmt.Sprintf(`{"linesOk":%d,"linesInvalid":0,"error":null}`, len(lines)))
		require.NoError(t, err)
	}))
	defer ts.Close()

	d := &Dynatrace{
		URL:               ts.URL,
		APIToken:          "123",
		Log:               testutil.Logger{},
		AddCounterMetrics: []string{},
	}

	err := d.Init()
	require.NoError(t, err)
	err = d.Connect()
	require.NoError(t, err)

	// Init metrics

	// Simple metrics are exported as a gauge unless in additional_counters
	expected = append(expected, "simple_metric.value,dt.metrics.source=telegraf gauge,3.14 1289430000000")
	expected = append(expected, "simple_metric.counter,dt.metrics.source=telegraf count,delta=5 1289430000000")
	d.AddCounterMetrics = append(d.AddCounterMetrics, "simple_metric.counter")
	m1 := metric.New(
		"simple_metric",
		map[string]string{},
		map[string]interface{}{"value": float64(3.14), "counter": 5},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	// Even if Type() returns counter, all metrics are treated as a gauge unless explicitly added to additional_counters
	expected = append(expected, "counter_type.value,dt.metrics.source=telegraf gauge,3.14 1289430000000")
	expected = append(expected, "counter_type.counter,dt.metrics.source=telegraf count,delta=5 1289430000000")
	d.AddCounterMetrics = append(d.AddCounterMetrics, "counter_type.counter")
	m2 := metric.New(
		"counter_type",
		map[string]string{},
		map[string]interface{}{"value": float64(3.14), "counter": 5},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
		telegraf.Counter,
	)

	expected = append(expected, "complex_metric.int,dt.metrics.source=telegraf gauge,1 1289430000000")
	expected = append(expected, "complex_metric.int64,dt.metrics.source=telegraf gauge,2 1289430000000")
	expected = append(expected, "complex_metric.float,dt.metrics.source=telegraf gauge,3 1289430000000")
	expected = append(expected, "complex_metric.float64,dt.metrics.source=telegraf gauge,4 1289430000000")
	expected = append(expected, "complex_metric.true,dt.metrics.source=telegraf gauge,1 1289430000000")
	expected = append(expected, "complex_metric.false,dt.metrics.source=telegraf gauge,0 1289430000000")
	m3 := metric.New(
		"complex_metric",
		map[string]string{},
		map[string]interface{}{"int": 1, "int64": int64(2), "float": 3.0, "float64": float64(4.0), "true": true, "false": false},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics := []telegraf.Metric{m1, m2, m3}

	err = d.Write(metrics)
	require.NoError(t, err)
}

func TestSendSingleMetricWithUnorderedTags(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check the encoded result
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		bodyString := string(bodyBytes)
		// use regex because dimension order isn't guaranteed
		require.Equal(t, len(bodyString), 94)
		require.Regexp(t, regexp.MustCompile(`^mymeasurement\.myfield`), bodyString)
		require.Regexp(t, regexp.MustCompile(`a=test`), bodyString)
		require.Regexp(t, regexp.MustCompile(`b=test`), bodyString)
		require.Regexp(t, regexp.MustCompile(`c=test`), bodyString)
		require.Regexp(t, regexp.MustCompile(`dt.metrics.source=telegraf`), bodyString)
		require.Regexp(t, regexp.MustCompile(`gauge,3.14 1289430000000$`), bodyString)
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(`{"linesOk":1,"linesInvalid":0,"error":null}`)
		require.NoError(t, err)
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
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		bodyString := string(bodyBytes)
		expected := "mymeasurement.myfield,dt.metrics.source=telegraf gauge,3.14 1289430000000"
		if bodyString != expected {
			t.Errorf("Metric encoding failed. expected: %#v but got: %#v", expected, bodyString)
		}
		err = json.NewEncoder(w).Encode(`{"linesOk":1,"linesInvalid":0,"error":null}`)
		require.NoError(t, err)
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
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		bodyString := string(bodyBytes)

		// use regex because dimension order isn't guaranteed
		require.Equal(t, len(bodyString), 100)
		require.Regexp(t, regexp.MustCompile(`^mymeasurement\.myfield`), bodyString)
		require.Regexp(t, regexp.MustCompile(`aaa=test`), bodyString)
		require.Regexp(t, regexp.MustCompile(`b_b=test`), bodyString)
		require.Regexp(t, regexp.MustCompile(`ccc=test`), bodyString)
		require.Regexp(t, regexp.MustCompile(`dt.metrics.source=telegraf`), bodyString)
		require.Regexp(t, regexp.MustCompile(`gauge,3.14 1289430000000$`), bodyString)

		err = json.NewEncoder(w).Encode(`{"linesOk":1,"linesInvalid":0,"error":null}`)
		require.NoError(t, err)
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
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		bodyString := string(bodyBytes)
		// use regex because field order isn't guaranteed
		require.Equal(t, len(bodyString), 132)
		require.Contains(t, bodyString, "mymeasurement.yes,dt.metrics.source=telegraf gauge,1 1289430000000")
		require.Contains(t, bodyString, "mymeasurement.no,dt.metrics.source=telegraf gauge,0 1289430000000")
		err = json.NewEncoder(w).Encode(`{"linesOk":1,"linesInvalid":0,"error":null}`)
		require.NoError(t, err)
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

func TestSendMetricWithDefaultDimensions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// check the encoded result
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		bodyString := string(bodyBytes)
		// use regex because field order isn't guaranteed
		require.Equal(t, len(bodyString), 78)
		require.Regexp(t, regexp.MustCompile("^mymeasurement.value"), bodyString)
		require.Regexp(t, regexp.MustCompile("dt.metrics.source=telegraf"), bodyString)
		require.Regexp(t, regexp.MustCompile("dim=value"), bodyString)
		require.Regexp(t, regexp.MustCompile("gauge,2 1289430000000$"), bodyString)
		err = json.NewEncoder(w).Encode(`{"linesOk":1,"linesInvalid":0,"error":null}`)
		require.NoError(t, err)
	}))
	defer ts.Close()

	d := &Dynatrace{DefaultDimensions: map[string]string{"dim": "value"}}

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
		map[string]interface{}{"value": 2},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics := []telegraf.Metric{m1}

	err = d.Write(metrics)
	require.NoError(t, err)
}

func TestMetricDimensionsOverrideDefault(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// check the encoded result
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		bodyString := string(bodyBytes)
		// use regex because field order isn't guaranteed
		require.Equal(t, len(bodyString), 80)
		require.Regexp(t, regexp.MustCompile("^mymeasurement.value"), bodyString)
		require.Regexp(t, regexp.MustCompile("dt.metrics.source=telegraf"), bodyString)
		require.Regexp(t, regexp.MustCompile("dim=metric"), bodyString)
		require.Regexp(t, regexp.MustCompile("gauge,32 1289430000000$"), bodyString)
		err = json.NewEncoder(w).Encode(`{"linesOk":1,"linesInvalid":0,"error":null}`)
		require.NoError(t, err)
	}))
	defer ts.Close()

	d := &Dynatrace{DefaultDimensions: map[string]string{"dim": "default"}}

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
		map[string]string{"dim": "metric"},
		map[string]interface{}{"value": 32},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics := []telegraf.Metric{m1}

	err = d.Write(metrics)
	require.NoError(t, err)
}

func TestStaticDimensionsOverrideMetric(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// check the encoded result
		bodyBytes, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		bodyString := string(bodyBytes)
		// use regex because field order isn't guaranteed
		require.Equal(t, len(bodyString), 53)
		require.Regexp(t, regexp.MustCompile("^mymeasurement.value"), bodyString)
		require.Regexp(t, regexp.MustCompile("dim=static"), bodyString)
		require.Regexp(t, regexp.MustCompile("gauge,32 1289430000000$"), bodyString)
		err = json.NewEncoder(w).Encode(`{"linesOk":1,"linesInvalid":0,"error":null}`)
		require.NoError(t, err)
	}))
	defer ts.Close()

	d := &Dynatrace{DefaultDimensions: map[string]string{"dim": "default"}}

	d.URL = ts.URL
	d.APIToken = "123"
	d.Log = testutil.Logger{}
	err := d.Init()
	require.NoError(t, err)
	err = d.Connect()
	require.NoError(t, err)

	d.normalizedStaticDimensions = dimensions.NewNormalizedDimensionList(dimensions.NewDimension("dim", "static"))

	// Init metrics

	m1 := metric.New(
		"mymeasurement",
		map[string]string{"dim": "metric"},
		map[string]interface{}{"value": 32},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics := []telegraf.Metric{m1}

	err = d.Write(metrics)
	require.NoError(t, err)
}

var warnfCalledTimes int

type loggerStub struct {
	testutil.Logger
}

func (l loggerStub) Warnf(format string, args ...interface{}) {
	warnfCalledTimes++
}

func TestSendUnsupportedMetric(t *testing.T) {
	warnfCalledTimes = 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not export because the only metric is an invalid type")
	}))
	defer ts.Close()

	d := &Dynatrace{}

	logStub := loggerStub{}

	d.URL = ts.URL
	d.APIToken = "123"
	d.Log = logStub
	err := d.Init()
	require.NoError(t, err)
	err = d.Connect()
	require.NoError(t, err)

	// Init metrics

	m1 := metric.New(
		"mymeasurement",
		map[string]string{},
		map[string]interface{}{"metric1": "unsupported_type"},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics := []telegraf.Metric{m1}

	err = d.Write(metrics)
	require.NoError(t, err)
	// Warnf called for invalid export
	require.Equal(t, 1, warnfCalledTimes)

	err = d.Write(metrics)
	require.NoError(t, err)
	// Warnf skipped for more invalid exports with the same name
	require.Equal(t, 1, warnfCalledTimes)

	m2 := metric.New(
		"mymeasurement",
		map[string]string{},
		map[string]interface{}{"metric2": "unsupported_type"},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	metrics = []telegraf.Metric{m2}

	err = d.Write(metrics)
	require.NoError(t, err)
	// Warnf called again for invalid export with a new metric name
	require.Equal(t, 2, warnfCalledTimes)

	err = d.Write(metrics)
	require.NoError(t, err)
	// Warnf skipped for more invalid exports with the same name
	require.Equal(t, 2, warnfCalledTimes)
}
