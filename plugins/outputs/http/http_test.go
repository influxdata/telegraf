package http

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/stretchr/testify/require"
)

func getMetric() telegraf.Metric {
	m, err := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	if err != nil {
		panic(err)
	}
	return m
}

func TestInvalidMethod(t *testing.T) {
	plugin := &HTTP{
		URL:    "",
		Method: http.MethodGet,
	}

	err := plugin.Connect()
	require.Error(t, err)
}

func TestMethod(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name           string
		plugin         *HTTP
		expectedMethod string
		connectError   bool
	}{
		{
			name: "default method is POST",
			plugin: &HTTP{
				URL:    u.String(),
				Method: defaultMethod,
			},
			expectedMethod: http.MethodPost,
		},
		{
			name: "put is okay",
			plugin: &HTTP{
				URL:    u.String(),
				Method: http.MethodPut,
			},
			expectedMethod: http.MethodPut,
		},
		{
			name: "get is invalid",
			plugin: &HTTP{
				URL:    u.String(),
				Method: http.MethodGet,
			},
			connectError: true,
		},
		{
			name: "method is case insensitive",
			plugin: &HTTP{
				URL:    u.String(),
				Method: "poST",
			},
			expectedMethod: http.MethodPost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.expectedMethod, r.Method)
				w.WriteHeader(http.StatusOK)
			})

			serializer := influx.NewSerializer()
			tt.plugin.SetSerializer(serializer)
			err = tt.plugin.Connect()
			if tt.connectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			err = tt.plugin.Write([]telegraf.Metric{getMetric()})
			require.NoError(t, err)
		})
	}
}

func TestStatusCode(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name       string
		plugin     *HTTP
		statusCode int
		errFunc    func(t *testing.T, err error)
	}{
		{
			name: "success",
			plugin: &HTTP{
				URL: u.String(),
			},
			statusCode: http.StatusOK,
			errFunc: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "1xx status is an error",
			plugin: &HTTP{
				URL: u.String(),
			},
			statusCode: 103,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "3xx status is an error",
			plugin: &HTTP{
				URL: u.String(),
			},
			statusCode: http.StatusMultipleChoices,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "4xx status is an error",
			plugin: &HTTP{
				URL: u.String(),
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

			serializer := influx.NewSerializer()
			tt.plugin.SetSerializer(serializer)
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
		plugin   *HTTP
		expected string
	}{
		{
			name: "default is text plain",
			plugin: &HTTP{
				URL: u.String(),
			},
			expected: defaultContentType,
		},
		{
			name: "overwrite content_type",
			plugin: &HTTP{
				URL:     u.String(),
				Headers: map[string]string{"Content-Type": "application/json"},
			},
			expected: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, tt.expected, r.Header.Get("Content-Type"))
				w.WriteHeader(http.StatusOK)
			})

			serializer := influx.NewSerializer()
			tt.plugin.SetSerializer(serializer)
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
		plugin   *HTTP
		payload  string
		expected string
	}{
		{
			name: "default is no content encoding",
			plugin: &HTTP{
				URL: u.String(),
			},
			expected: "",
		},
		{
			name: "overwrite content_encoding",
			plugin: &HTTP{
				URL:             u.String(),
				ContentEncoding: "gzip",
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
				require.Contains(t, string(payload), "cpu value=42")

				w.WriteHeader(http.StatusNoContent)
			})

			serializer := influx.NewSerializer()
			tt.plugin.SetSerializer(serializer)
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
		plugin *HTTP
	}{
		{
			name: "default",
			plugin: &HTTP{
				URL: u.String(),
			},
		},
		{
			name: "username only",
			plugin: &HTTP{
				URL:      u.String(),
				Username: "username",
			},
		},
		{
			name: "password only",
			plugin: &HTTP{
				URL:      u.String(),
				Password: "pa$$word",
			},
		},
		{
			name: "username and password",
			plugin: &HTTP{
				URL:      u.String(),
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

			serializer := influx.NewSerializer()
			tt.plugin.SetSerializer(serializer)
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
		plugin       *HTTP
		tokenHandler TestHandlerFunc
		handler      TestHandlerFunc
	}{
		{
			name: "no credentials",
			plugin: &HTTP{
				URL: u.String(),
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Len(t, r.Header["Authorization"], 0)
				w.WriteHeader(http.StatusOK)
			},
		},
		{
			name: "success",
			plugin: &HTTP{
				URL:          u.String() + "/write",
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
				case "/write":
					tt.handler(t, w, r)
				case "/token":
					tt.tokenHandler(t, w, r)
				}
			})

			serializer := influx.NewSerializer()
			tt.plugin.SetSerializer(serializer)
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

	internal.SetVersion("1.2.3")

	t.Run("default-user-agent", func(t *testing.T) {
		ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "Telegraf/1.2.3", r.Header.Get("User-Agent"))
			w.WriteHeader(http.StatusOK)
		})

		client := &HTTP{
			URL:    u.String(),
			Method: defaultMethod,
		}

		serializer := influx.NewSerializer()
		client.SetSerializer(serializer)
		err = client.Connect()
		require.NoError(t, err)

		err = client.Write([]telegraf.Metric{getMetric()})
		require.NoError(t, err)
	})
}
