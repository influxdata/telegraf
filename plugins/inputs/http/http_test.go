package http_test

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
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

func TestBodyAndContentEncoding(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	url := fmt.Sprintf("http://%s", ts.Listener.Addr().String())

	tests := []struct {
		name             string
		plugin           *plugin.HTTP
		queryHandlerFunc func(t *testing.T, w http.ResponseWriter, r *http.Request)
	}{
		{
			name: "no body",
			plugin: &plugin.HTTP{
				Method: "POST",
				URLs:   []string{url},
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, []byte(""), body)
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "post body",
			plugin: &plugin.HTTP{
				URLs:   []string{url},
				Method: "POST",
				Body:   "test",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, []byte("test"), body)
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "get method body is sent",
			plugin: &plugin.HTTP{
				URLs:   []string{url},
				Method: "GET",
				Body:   "test",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, []byte("test"), body)
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "gzip encoding",
			plugin: &plugin.HTTP{
				URLs:            []string{url},
				Method:          "GET",
				Body:            "test",
				ContentEncoding: "gzip",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, r.Header.Get("Content-Encoding"), "gzip")

				gr, err := gzip.NewReader(r.Body)
				require.NoError(t, err)
				body, err := ioutil.ReadAll(gr)
				require.NoError(t, err)
				require.Equal(t, []byte("test"), body)
				w.WriteHeader(http.StatusOK)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.queryHandlerFunc(t, w, r)
			})

			parser, err := parsers.NewParser(&parsers.Config{DataFormat: "influx"})
			require.NoError(t, err)

			tt.plugin.SetParser(parser)

			var acc testutil.Accumulator
			err = tt.plugin.Gather(&acc)
			require.NoError(t, err)
		})
	}
}
