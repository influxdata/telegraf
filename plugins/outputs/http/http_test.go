package http

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/influxdata/telegraf/config"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	internalaws "github.com/influxdata/telegraf/config/aws"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	httpconfig "github.com/influxdata/telegraf/plugins/common/http"
	"github.com/influxdata/telegraf/plugins/common/oauth"
	"github.com/influxdata/telegraf/plugins/serializers"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/plugins/serializers/json"
	"github.com/influxdata/telegraf/testutil"
)

func getMetric() telegraf.Metric {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)

	return m
}

func getMetrics(n int) []telegraf.Metric {
	m := make([]telegraf.Metric, n)
	for n > 0 {
		n--
		m[n] = getMetric()
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

func TestHTTPClientConfig(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name                        string
		plugin                      *HTTP
		connectError                bool
		expectedMaxIdleConns        int
		expectedMaxIdleConnsPerHost int
	}{
		{
			name: "With default client Config",
			plugin: &HTTP{
				URL:    u.String(),
				Method: defaultMethod,
				HTTPClientConfig: httpconfig.HTTPClientConfig{
					IdleConnTimeout: config.Duration(5 * time.Second),
				},
			},
			expectedMaxIdleConns:        0,
			expectedMaxIdleConnsPerHost: 0,
		},
		{
			name: "With MaxIdleConns client Config",
			plugin: &HTTP{
				URL:    u.String(),
				Method: defaultMethod,
				HTTPClientConfig: httpconfig.HTTPClientConfig{
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 100,
					IdleConnTimeout:     config.Duration(5 * time.Second),
				},
			},
			expectedMaxIdleConns:        100,
			expectedMaxIdleConnsPerHost: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			tr := tt.plugin.client.Transport.(*http.Transport)
			maxIdleConns, maxIdleConnsPerHost := tr.MaxIdleConns, tr.MaxIdleConnsPerHost
			require.Equal(t, tt.expectedMaxIdleConns, maxIdleConns)
			require.Equal(t, tt.expectedMaxIdleConnsPerHost, maxIdleConnsPerHost)

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
			statusCode: http.StatusSwitchingProtocols,
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
			statusCode: http.StatusBadRequest,
			errFunc: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "Do not retry on configured non-retryable statuscode",
			plugin: &HTTP{
				URL:                     u.String(),
				NonRetryableStatusCodes: []int{409},
			},
			statusCode: http.StatusConflict,
			errFunc: func(t *testing.T, err error) {
				require.NoError(t, err)
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

			tt.plugin.Log = testutil.Logger{}

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

				payload, err := io.ReadAll(body)
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
				URL: u.String() + "/write",
				HTTPClientConfig: httpconfig.HTTPClientConfig{
					OAuth2Config: oauth.OAuth2Config{
						ClientID:     "howdy",
						ClientSecret: "secret",
						TokenURL:     u.String() + "/token",
						Scopes:       []string{"urn:opc:idm:__myscopes__"},
					},
				},
			},
			tokenHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				values := url.Values{}
				values.Add("access_token", token)
				values.Add("token_type", "bearer")
				values.Add("expires_in", "3600")
				_, err = w.Write([]byte(values.Encode()))
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

	t.Run("default-user-agent", func(t *testing.T) {
		ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, internal.ProductToken(), r.Header.Get("User-Agent"))
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

func TestBatchedUnbatched(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	client := &HTTP{
		URL:    u.String(),
		Method: defaultMethod,
	}

	var s = map[string]serializers.Serializer{
		"influx": influx.NewSerializer(),
		"json": func(s serializers.Serializer, err error) serializers.Serializer {
			require.NoError(t, err)
			return s
		}(json.NewSerializer(time.Second, "")),
	}

	for name, serializer := range s {
		var requests int
		ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requests++
			w.WriteHeader(http.StatusOK)
		})

		t.Run(name, func(t *testing.T) {
			for _, mode := range [...]bool{false, true} {
				requests = 0
				client.UseBatchFormat = mode
				client.SetSerializer(serializer)

				err = client.Connect()
				require.NoError(t, err)
				err = client.Write(getMetrics(3))
				require.NoError(t, err)

				if client.UseBatchFormat {
					require.Equal(t, requests, 1, "batched")
				} else {
					require.Equal(t, requests, 3, "unbatched")
				}
			}
		})
	}
}

func TestAwsCredentials(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name         string
		plugin       *HTTP
		tokenHandler TestHandlerFunc
		handler      TestHandlerFunc
	}{
		{
			name: "simple credentials",
			plugin: &HTTP{
				URL:        u.String(),
				AwsService: "aps",
				CredentialConfig: internalaws.CredentialConfig{
					Region:    "us-east-1",
					AccessKey: "dummy",
					SecretKey: "dummy",
				},
			},
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Contains(t, r.Header["Authorization"][0], "AWS4-HMAC-SHA256")
				require.Contains(t, r.Header["Authorization"][0], "=dummy/")
				require.Contains(t, r.Header["Authorization"][0], "/us-east-1/")
				w.WriteHeader(http.StatusOK)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.handler(t, w, r)
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
