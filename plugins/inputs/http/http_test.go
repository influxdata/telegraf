package http_test

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/common/oauth"
	httpplugin "github.com/influxdata/telegraf/plugins/inputs/http"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/plugins/parsers/csv"
	"github.com/influxdata/telegraf/testutil"
)

func TestHTTPWithJSONFormat(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, _ = w.Write([]byte(simpleJSON))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	address := fakeServer.URL + "/endpoint"
	plugin := &httpplugin.HTTP{
		URLs: []string{address},
		Log:  testutil.Logger{},
	}
	metricName := "metricName"

	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		return parsers.NewParser(&parsers.Config{
			DataFormat: "json",
			MetricName: "metricName",
		})
	})

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, acc.GatherError(plugin.Gather))

	require.Len(t, acc.Metrics, 1)

	// basic check to see if we got the right field, value and tag
	var metric = acc.Metrics[0]
	require.Equal(t, metric.Measurement, metricName)
	require.Len(t, acc.Metrics[0].Fields, 1)
	require.Equal(t, acc.Metrics[0].Fields["a"], 1.2)
	require.Equal(t, acc.Metrics[0].Tags["url"], address)
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

	address := fakeServer.URL + "/endpoint"
	plugin := &httpplugin.HTTP{
		URLs:    []string{address},
		Headers: map[string]string{header: headerValue},
		Log:     testutil.Logger{},
	}

	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		return parsers.NewParser(&parsers.Config{
			DataFormat: "json",
			MetricName: "metricName",
		})
	})

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, acc.GatherError(plugin.Gather))
}

func TestInvalidStatusCode(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer fakeServer.Close()

	address := fakeServer.URL + "/endpoint"
	plugin := &httpplugin.HTTP{
		URLs: []string{address},
		Log:  testutil.Logger{},
	}

	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		return parsers.NewParser(&parsers.Config{
			DataFormat: "json",
			MetricName: "metricName",
		})
	})

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.Error(t, acc.GatherError(plugin.Gather))
}

func TestSuccessStatusCodes(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer fakeServer.Close()

	address := fakeServer.URL + "/endpoint"
	plugin := &httpplugin.HTTP{
		URLs:               []string{address},
		SuccessStatusCodes: []int{200, 202},
		Log:                testutil.Logger{},
	}

	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		return parsers.NewParser(&parsers.Config{
			DataFormat: "json",
			MetricName: "metricName",
		})
	})

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, acc.GatherError(plugin.Gather))
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

	plugin := &httpplugin.HTTP{
		URLs:   []string{fakeServer.URL},
		Method: "POST",
		Log:    testutil.Logger{},
	}

	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		return parsers.NewParser(&parsers.Config{
			DataFormat: "json",
			MetricName: "metricName",
		})
	})

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, acc.GatherError(plugin.Gather))
}

const simpleJSON = `
{
    "a": 1.2
}
`
const simpleCSVWithHeader = `
# Simple CSV with header(s)
a,b,c
1.2,3.1415,ok
`

func TestBodyAndContentEncoding(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	address := fmt.Sprintf("http://%s", ts.Listener.Addr().String())

	tests := []struct {
		name             string
		plugin           *httpplugin.HTTP
		queryHandlerFunc func(t *testing.T, w http.ResponseWriter, r *http.Request)
	}{
		{
			name: "no body",
			plugin: &httpplugin.HTTP{
				Method: "POST",
				URLs:   []string{address},
				Log:    testutil.Logger{},
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, []byte(""), body)
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "post body",
			plugin: &httpplugin.HTTP{
				URLs:   []string{address},
				Method: "POST",
				Body:   "test",
				Log:    testutil.Logger{},
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, []byte("test"), body)
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "get method body is sent",
			plugin: &httpplugin.HTTP{
				URLs:   []string{address},
				Method: "GET",
				Body:   "test",
				Log:    testutil.Logger{},
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, []byte("test"), body)
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "gzip encoding",
			plugin: &httpplugin.HTTP{
				URLs:            []string{address},
				Method:          "GET",
				Body:            "test",
				ContentEncoding: "gzip",
				Log:             testutil.Logger{},
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, r.Header.Get("Content-Encoding"), "gzip")

				gr, err := gzip.NewReader(r.Body)
				require.NoError(t, err)
				body, err := io.ReadAll(gr)
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

			tt.plugin.SetParserFunc(func() (telegraf.Parser, error) {
				return parsers.NewParser(&parsers.Config{DataFormat: "influx"})
			})

			var acc testutil.Accumulator
			require.NoError(t, tt.plugin.Init())
			require.NoError(t, tt.plugin.Gather(&acc))
		})
	}
}

type TestHandlerFunc func(t *testing.T, w http.ResponseWriter, r *http.Request)

func TestOAuthClientCredentialsGrant(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	var token = "2YotnFZFEjr1zCsicMWpAA"

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name         string
		plugin       *httpplugin.HTTP
		tokenHandler TestHandlerFunc
		handler      TestHandlerFunc
	}{
		{
			name: "no credentials",
			plugin: &httpplugin.HTTP{
				URLs: []string{u.String()},
				Log:  testutil.Logger{},
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Len(t, r.Header["Authorization"], 0)
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "success",
			plugin: &httpplugin.HTTP{
				URLs: []string{u.String() + "/write"},
				HTTPClientConfig: httpconfig.HTTPClientConfig{
					OAuth2Config: oauth.OAuth2Config{
						ClientID:     "howdy",
						ClientSecret: "secret",
						TokenURL:     u.String() + "/token",
						Scopes:       []string{"urn:opc:idm:__myscopes__"},
					},
				},
				Log: testutil.Logger{},
			},
			tokenHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				values := url.Values{}
				values.Add("access_token", token)
				values.Add("token_type", "bearer")
				values.Add("expires_in", "3600")
				_, err := w.Write([]byte(values.Encode()))
				require.NoError(t, err)
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, []string{"Bearer " + token}, r.Header["Authorization"])
				w.WriteHeader(http.StatusOK)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/write":
					tt.handler(t, w, r)
				case "/token":
					tt.tokenHandler(t, w, r)
				}
			})

			tt.plugin.SetParserFunc(func() (telegraf.Parser, error) {
				return parsers.NewValueParser("metric", "string", "", nil)
			})

			err = tt.plugin.Init()
			require.NoError(t, err)

			var acc testutil.Accumulator
			err = tt.plugin.Gather(&acc)
			require.NoError(t, err)
		})
	}
}

func TestHTTPWithCSVFormat(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			_, _ = w.Write([]byte(simpleCSVWithHeader))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	address := fakeServer.URL + "/endpoint"
	plugin := &httpplugin.HTTP{
		URLs: []string{address},
		Log:  testutil.Logger{},
	}

	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		parser := &csv.Parser{
			MetricName:  "metricName",
			SkipRows:    3,
			ColumnNames: []string{"a", "b", "c"},
			TagColumns:  []string{"c"},
		}
		err := parser.Init()
		return parser, err
	})

	expected := []telegraf.Metric{
		testutil.MustMetric("metricName",
			map[string]string{
				"url": address,
				"c":   "ok",
			},
			map[string]interface{}{
				"a": 1.2,
				"b": 3.1415,
			},
			time.Unix(0, 0),
		),
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, acc.GatherError(plugin.Gather))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())

	// Run the parser a second time to test for correct stateful handling
	acc.ClearMetrics()
	require.NoError(t, acc.GatherError(plugin.Gather))
	testutil.RequireMetricsEqual(t, expected, acc.GetTelegrafMetrics(), testutil.IgnoreTime())
}
