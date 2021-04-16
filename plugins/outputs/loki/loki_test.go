package loki

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/influxdata/telegraf/testutil"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/stretchr/testify/require"
)

func getMetric() telegraf.Metric {
	return testutil.MustMetric(
		"log",
		map[string]string{
			"key1": "value1",
		},
		map[string]interface{}{
			"line":  "my log",
			"field": 3.14,
		},
		time.Unix(123, 0),
	)
}

func TestStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name       string
		plugin     *Loki
		statusCode int
		errFunc    func(t *testing.T, err error)
	}{
		{
			name: "success",
			plugin: &Loki{
				Domain: u.String(),
			},
			statusCode: http.StatusNoContent,
			errFunc: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "1xx status is an error",
			plugin: &Loki{
				Domain: u.String(),
			},
			statusCode: 103,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "3xx status is an error",
			plugin: &Loki{
				Domain: u.String(),
			},
			statusCode: http.StatusMultipleChoices,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "4xx status is an error",
			plugin: &Loki{
				Domain: u.String(),
			},
			statusCode: http.StatusMultipleChoices,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			err = tt.plugin.Connect()
			require.NoError(t, err)

			err = tt.plugin.Write([]telegraf.Metric{getMetric()})
			tt.errFunc(t, err)
		})
	}
}

func TestContentType(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name     string
		plugin   *Loki
		expected string
	}{
		{
			name: "default is application/json",
			plugin: &Loki{
				Domain: u.String(),
			},
			expected: "application/json",
		},
		{
			name: "overwrite content_type",
			plugin: &Loki{
				Domain:  u.String(),
				Headers: map[string]string{"Content-Type": "plain/text"},
			},
			// plugin force content-type
			expected: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.expected, r.Header.Get("Content-Type"))
				w.WriteHeader(http.StatusOK)
			})

			err = tt.plugin.Connect()
			require.NoError(t, err)

			err = tt.plugin.Write([]telegraf.Metric{getMetric()})
			require.NoError(t, err)
		})
	}
}

func TestContentEncodingGzip(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name     string
		plugin   *Loki
		expected string
	}{
		{
			name: "default is no content encoding",
			plugin: &Loki{
				Domain: u.String(),
			},
			expected: "",
		},
		{
			name: "overwrite content_encoding",
			plugin: &Loki{
				Domain:      u.String(),
				GZipRequest: true,
			},
			expected: "gzip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.expected, r.Header.Get("Content-Encoding"))

				body := r.Body
				var err error
				if r.Header.Get("Content-Encoding") == "gzip" {
					body, err = gzip.NewReader(r.Body)
					require.NoError(t, err)
				}

				payload, err := ioutil.ReadAll(body)
				require.NoError(t, err)

				var s Request
				err = json.Unmarshal(payload, &s)
				require.NoError(t, err)
				require.Len(t, s.Streams, 1)
				require.Len(t, s.Streams[0].Logs, 1)
				require.Len(t, s.Streams[0].Logs[0], 2)
				require.Equal(t, map[string]string{"key1": "value1"}, s.Streams[0].Labels)
				require.Equal(t, "123000000000", s.Streams[0].Logs[0][0])
				require.Contains(t, s.Streams[0].Logs[0][1], "line=\"my log\"")
				require.Contains(t, s.Streams[0].Logs[0][1], "field=\"3.14\"")

				w.WriteHeader(http.StatusNoContent)
			})

			err = tt.plugin.Connect()
			require.NoError(t, err)

			err = tt.plugin.Write([]telegraf.Metric{getMetric()})
			require.NoError(t, err)
		})
	}
}

func TestBasicAuth(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name   string
		plugin *Loki
	}{
		{
			name: "default",
			plugin: &Loki{
				Domain: u.String(),
			},
		},
		{
			name: "username and password",
			plugin: &Loki{
				Domain:   u.String(),
				Username: "username",
				Password: "pa$$word",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				username, password, _ := r.BasicAuth()
				require.Equal(t, tt.plugin.Username, username)
				require.Equal(t, tt.plugin.Password, password)
				w.WriteHeader(http.StatusOK)
			})

			err = tt.plugin.Connect()
			require.NoError(t, err)

			err = tt.plugin.Write([]telegraf.Metric{getMetric()})
			require.NoError(t, err)
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
		plugin       *Loki
		tokenHandler TestHandlerFunc
		handler      TestHandlerFunc
	}{
		{
			name: "no credentials",
			plugin: &Loki{
				Domain: u.String(),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Len(t, r.Header["Authorization"], 0)
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "success",
			plugin: &Loki{
				Domain:       u.String(),
				ClientID:     "howdy",
				ClientSecret: "secret",
				TokenURL:     u.String() + "/token",
				Scopes:       []string{"urn:opc:idm:__myscopes__"},
			},
			tokenHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				values := url.Values{}
				values.Add("access_token", token)
				values.Add("token_type", "bearer")
				values.Add("expires_in", "3600")
				w.Write([]byte(values.Encode()))
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
				case defaultEndpoint:
					tt.handler(t, w, r)
				case "/token":
					tt.tokenHandler(t, w, r)
				}
			})

			err = tt.plugin.Connect()
			require.NoError(t, err)

			err = tt.plugin.Write([]telegraf.Metric{getMetric()})
			require.NoError(t, err)
		})
	}
}

func TestDefaultUserAgent(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	t.Run("default-user-agent", func(t *testing.T) {
		ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, internal.ProductToken(), r.Header.Get("User-Agent"))
			w.WriteHeader(http.StatusOK)
		})

		client := &Loki{
			Domain: u.String(),
		}

		err = client.Connect()
		require.NoError(t, err)

		err = client.Write([]telegraf.Metric{getMetric()})
		require.NoError(t, err)
	})
}
