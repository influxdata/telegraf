package http_test

import (
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	common_http "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/common/oauth"
	httpplugin "github.com/influxdata/telegraf/plugins/inputs/http"
	"github.com/influxdata/telegraf/plugins/parsers/csv"
	"github.com/influxdata/telegraf/plugins/parsers/influx"
	"github.com/influxdata/telegraf/plugins/parsers/json"
	"github.com/influxdata/telegraf/plugins/parsers/value"
	"github.com/influxdata/telegraf/testutil"
)

func TestHTTPWithJSONFormat(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			if _, err := w.Write([]byte(simpleJSON)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
				return
			}
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
		p := &json.Parser{MetricName: "metricName"}
		err := p.Init()
		return p, err
	})

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, acc.GatherError(plugin.Gather))

	require.Len(t, acc.Metrics, 1)

	// basic check to see if we got the right field, value and tag
	var metric = acc.Metrics[0]
	require.Equal(t, metric.Measurement, metricName)
	require.Len(t, acc.Metrics[0].Fields, 1)
	require.InDelta(t, 1.2, acc.Metrics[0].Fields["a"], testutil.DefaultDelta)
	require.Equal(t, acc.Metrics[0].Tags["url"], address)
}

func TestHTTPHeaders(t *testing.T) {
	header := "X-Special-Header"
	headerValue := "Special-Value"
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			if r.Header.Get(header) == headerValue {
				if _, err := w.Write([]byte(simpleJSON)); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Error(err)
					return
				}
			} else {
				w.WriteHeader(http.StatusForbidden)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeServer.Close()

	address := fakeServer.URL + "/endpoint"
	headerSecret := config.NewSecret([]byte(headerValue))
	plugin := &httpplugin.HTTP{
		URLs:    []string{address},
		Headers: map[string]*config.Secret{header: &headerSecret},
		Log:     testutil.Logger{},
	}

	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		p := &json.Parser{MetricName: "metricName"}
		err := p.Init()
		return p, err
	})

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, acc.GatherError(plugin.Gather))
}

func TestHTTPContentLengthHeader(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endpoint" {
			if r.Header.Get("Content-Length") != "" {
				if _, err := w.Write([]byte(simpleJSON)); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					t.Error(err)
					return
				}
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
		Headers: map[string]*config.Secret{},
		Body:    "{}",
		Log:     testutil.Logger{},
	}

	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		p := &json.Parser{MetricName: "metricName"}
		err := p.Init()
		return p, err
	})

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.NoError(t, acc.GatherError(plugin.Gather))
}

func TestInvalidStatusCode(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer fakeServer.Close()

	address := fakeServer.URL + "/endpoint"
	plugin := &httpplugin.HTTP{
		URLs: []string{address},
		Log:  testutil.Logger{},
	}

	plugin.SetParserFunc(func() (telegraf.Parser, error) {
		p := &json.Parser{MetricName: "metricName"}
		err := p.Init()
		return p, err
	})

	var acc testutil.Accumulator
	require.NoError(t, plugin.Init())
	require.Error(t, acc.GatherError(plugin.Gather))
}

func TestSuccessStatusCodes(t *testing.T) {
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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
		p := &json.Parser{MetricName: "metricName"}
		err := p.Init()
		return p, err
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
		p := &json.Parser{MetricName: "metricName"}
		err := p.Init()
		return p, err
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

	address := "http://" + ts.Listener.Addr().String()

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
				require.Equal(t, "gzip", r.Header.Get("Content-Encoding"))

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
				parser := &influx.Parser{}
				err := parser.Init()
				return parser, err
			})

			var acc testutil.Accumulator
			require.NoError(t, tt.plugin.Init())
			require.NoError(t, tt.plugin.Gather(&acc))
		})
	}
}

type testHandlerFunc func(t *testing.T, w http.ResponseWriter, r *http.Request)

func TestOAuthClientCredentialsGrant(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	var token = "2YotnFZFEjr1zCsicMWpAA"

	u, err := url.Parse("http://" + ts.Listener.Addr().String())
	require.NoError(t, err)

	tests := []struct {
		name         string
		plugin       *httpplugin.HTTP
		tokenHandler testHandlerFunc
		handler      testHandlerFunc
	}{
		{
			name: "no credentials",
			plugin: &httpplugin.HTTP{
				URLs: []string{u.String()},
				Log:  testutil.Logger{},
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Empty(t, r.Header["Authorization"])
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "success",
			plugin: &httpplugin.HTTP{
				URLs: []string{u.String() + "/write"},
				HTTPClientConfig: common_http.HTTPClientConfig{
					OAuth2Config: oauth.OAuth2Config{
						ClientID:     "howdy",
						ClientSecret: "secret",
						TokenURL:     u.String() + "/token",
						Scopes:       []string{"urn:opc:idm:__myscopes__"},
					},
				},
				Log: testutil.Logger{},
			},
			tokenHandler: func(t *testing.T, w http.ResponseWriter, _ *http.Request) {
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
				p := &value.Parser{
					MetricName: "metric",
					DataType:   "string",
				}
				err := p.Init()
				return p, err
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
			if _, err := w.Write([]byte(simpleCSVWithHeader)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
				return
			}
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

const (
	httpOverUnixScheme = "http+unix"
)

func TestConnectionOverUnixSocket(t *testing.T) {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/data" {
			w.Header().Set("Content-Type", "text/csv")
			if _, err := w.Write([]byte(simpleCSVWithHeader)); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				t.Error(err)
				return
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	// The Maximum length of the socket path is 104/108 characters, path created with t.TempDir() is too long for some cases
	// (it combines test name with subtest name and some random numbers in the path). Therefore, in this case, it is safer to stick with `os.MkdirTemp()`.
	//nolint:usetesting // Ignore "os.TempDir() could be replaced by t.TempDir() in TestConnectionOverUnixSocket" finding.
	unixListenAddr := filepath.Join(os.TempDir(), fmt.Sprintf("httptestserver.%d.sock", rand.Intn(1_000_000)))
	t.Cleanup(func() { os.Remove(unixListenAddr) })

	unixListener, err := net.Listen("unix", unixListenAddr)
	require.NoError(t, err)

	ts.Listener = unixListener
	ts.Start()
	defer ts.Close()

	// NOTE: Remove ":" from windows filepath and replace all "\" with "/".
	//       This is *required* so that the unix socket path plays well with unixtransport.
	replacer := strings.NewReplacer(":", "", "\\", "/")
	sockPath := replacer.Replace(unixListenAddr)

	address := fmt.Sprintf("%s://%s:/data", httpOverUnixScheme, sockPath)
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
			time.Unix(22000, 0),
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
