package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	plugin "github.com/influxdata/telegraf/plugins/inputs/http"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestHTTPwithJSONFormat(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, _ = w.Write([]byte(simpleJSON))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	url := fakeServer.URL + "/endpoint"
	plugin := &plugin.HTTP{
		URLs: []string{url},
	}
	metricName := "metricName"

	p, _ := parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "metricName",
	})
	plugin.SetParser(p)

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))

	require.Len(t, acc.Metrics, 1)

	// basic check to see if we got the right field, value and tag
	var metric = acc.Metrics[0]
	require.Equal(t, metric.Measurement, metricName)
	require.Len(t, acc.Metrics[0].Fields, 1)
	require.Equal(t, acc.Metrics[0].Fields["a"], 1.2)
	require.Equal(t, acc.Metrics[0].Tags["url"], url)
}

func TestHTTPHeaders(t *testing.T) {
	header := "X-Special-Header"
	headerValue := "Special-Value"
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			if r.Header.Get(header) == headerValue {
				_, _ = w.Write([]byte(simpleJSON))
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	url := fakeServer.URL + "/endpoint"
	plugin := &plugin.HTTP{
		URLs:    []string{url},
		Headers: map[string]string{header: headerValue},
	}

	p, _ := parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "metricName",
	})
	plugin.SetParser(p)

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))
}

func TestInvalidStatusCode(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer fakeServer.Close()

	url := fakeServer.URL + "/endpoint"
	plugin := &plugin.HTTP{
		URLs: []string{url},
	}

	metricName := "metricName"
	p, _ := parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: metricName,
	})
	plugin.SetParser(p)

	var acc testutil.Accumulator
	require.Error(t, acc.GatherError(plugin.Gather))
}

func TestMethod(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	plugin := &plugin.HTTP{
		URLs:   []string{fakeServer.URL},
		Method: "POST",
	}

	p, _ := parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "metricName",
	})
	plugin.SetParser(p)

	var acc testutil.Accumulator
	require.NoError(t, acc.GatherError(plugin.Gather))
}

func TestParserNotSet(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, _ = w.Write([]byte(simpleJSON))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	url := fakeServer.URL + "/endpoint"
	plugin := &plugin.HTTP{
		URLs: []string{url},
	}

	var acc testutil.Accumulator
	require.Error(t, acc.GatherError(plugin.Gather))
}

const simpleJSON = `
{
    "a": 1.2
}
`
