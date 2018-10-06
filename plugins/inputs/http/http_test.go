package http_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	plugin "github.com/influxdata/telegraf/plugins/inputs/http"
	"github.com/influxdata/telegraf/plugins/parsers"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const simpleJSON = `
{
    "a": 1.2
}
`

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

func TestCookieJar(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name                string
		plugin              *plugin.HTTP
		cookieName          string
		cookieValue         string
		expiration          time.Time
		expectedCookieCount int
	}{
		{
			name: "cookie is set",
			plugin: &plugin.HTTP{
				URLs: []string{u.String()},
			},
			cookieName:          "Amaretti",
			cookieValue:         "Italy",
			expiration:          time.Now().Add(1 * time.Second),
			expectedCookieCount: 1,
		},
		{
			name: "expired cookie is not set",
			plugin: &plugin.HTTP{
				URLs: []string{u.String()},
			},
			cookieName:          "Macaron",
			cookieValue:         "France",
			expiration:          time.Now().Add(-1 * time.Second),
			expectedCookieCount: 0,
		},
	}

	parser, _ := parsers.NewParser(&parsers.Config{
		DataFormat: "json",
		MetricName: "metricName",
	})
	var acc testutil.Accumulator

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			firstReq := true
			ts.Config.Handler = http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					// Set a cookie when receiving the first request
					if firstReq {
						http.SetCookie(w, &http.Cookie{
							Name:    tt.cookieName,
							Value:   tt.cookieValue,
							Expires: tt.expiration,
						})
						firstReq = false
						return
					}

					// Check that subsequent requests contain the cookie set above
					cookies := r.Cookies()
					require.Len(t, cookies, tt.expectedCookieCount)

					if tt.expectedCookieCount > 0 {
						require.Equal(
							t,
							fmt.Sprintf("%s=%s", tt.cookieName, tt.cookieValue),
							cookies[0].String(),
						)
					}
				},
			)

			tt.plugin.SetParser(parser)

			// Send two requests to the mock server and check that the
			// cookies are propagated properly during the second request.
			for i := 0; i < 2; i++ {
				err = acc.GatherError(tt.plugin.Gather)
				require.NoError(t, err)
			}
		})
	}
}
