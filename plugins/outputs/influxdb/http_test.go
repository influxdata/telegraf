package influxdb_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/plugins/outputs/influxdb"
	"github.com/stretchr/testify/require"
)

func getHTTPURL() *url.URL {
	u, err := url.Parse("http://localhost")
	if err != nil {
		panic(err)
	}
	return u
}

func TestHTTP_EmptyConfig(t *testing.T) {
	config := &influxdb.HTTPConfig{}
	_, err := influxdb.NewHTTPClient(config)
	require.Error(t, err)
	require.Contains(t, err.Error(), influxdb.ErrMissingURL.Error())
}

func TestHTTP_MinimalConfig(t *testing.T) {
	config := &influxdb.HTTPConfig{
		URL: getHTTPURL(),
	}
	_, err := influxdb.NewHTTPClient(config)
	require.NoError(t, err)
}

func TestHTTP_UnsupportedScheme(t *testing.T) {
	config := &influxdb.HTTPConfig{
		URL: &url.URL{
			Scheme: "foo",
			Host:   "localhost",
		},
	}
	_, err := influxdb.NewHTTPClient(config)
	require.Error(t, err)
}

func TestHTTP_CreateDatabase(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	successResponse := []byte(`{"results": [{"statement_id": 0}]}`)

	tests := []struct {
		name             string
		config           *influxdb.HTTPConfig
		database         string
		queryHandlerFunc func(t *testing.T, w http.ResponseWriter, r *http.Request)
		errFunc          func(t *testing.T, err error)
	}{
		{
			name: "success",
			config: &influxdb.HTTPConfig{
				URL:      u,
				Database: "xyzzy",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, `CREATE DATABASE "xyzzy"`, r.FormValue("q"))
				w.WriteHeader(http.StatusOK)
				w.Write(successResponse)
			},
		},
		{
			name: "send basic auth",
			config: &influxdb.HTTPConfig{
				URL:      u,
				Username: "guy",
				Password: "smiley",
				Database: "telegraf",
			},
			database: "telegraf",
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				username, password, ok := r.BasicAuth()
				require.True(t, ok)
				require.Equal(t, "guy", username)
				require.Equal(t, "smiley", password)
				w.WriteHeader(http.StatusOK)
				w.Write(successResponse)
			},
		},
		{
			name: "send user agent",
			config: &influxdb.HTTPConfig{
				URL: u,
				Headers: map[string]string{
					"A": "B",
					"C": "D",
				},
				Database: "telegraf",
			},
			database: `a " b`,
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, r.Header.Get("A"), "B")
				require.Equal(t, r.Header.Get("C"), "D")
				w.WriteHeader(http.StatusOK)
				w.Write(successResponse)
			},
		},
		{
			name: "send headers",
			config: &influxdb.HTTPConfig{
				URL: u,
				Headers: map[string]string{
					"A": "B",
					"C": "D",
				},
				Database: "telegraf",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, r.Header.Get("A"), "B")
				require.Equal(t, r.Header.Get("C"), "D")
				w.WriteHeader(http.StatusOK)
				w.Write(successResponse)
			},
		},
		{
			name: "database default",
			config: &influxdb.HTTPConfig{
				URL: u,
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, `CREATE DATABASE "telegraf"`, r.FormValue("q"))
				w.WriteHeader(http.StatusOK)
				w.Write(successResponse)
			},
		},
		{
			name: "database name is escaped",
			config: &influxdb.HTTPConfig{
				URL:      u,
				Database: `a " b`,
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, `CREATE DATABASE "a \" b"`, r.FormValue("q"))
				w.WriteHeader(http.StatusOK)
				w.Write(successResponse)
			},
		},
		{
			name: "invalid database name creates api error",
			config: &influxdb.HTTPConfig{
				URL:      u,
				Database: `a \\ b`,
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				// Yes, 200 OK is the correct response...
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"results": [{"error": "invalid name", "statement_id": 0}]}`))
			},
			errFunc: func(t *testing.T, err error) {
				expected := &influxdb.APIError{
					StatusCode:  200,
					Title:       "200 OK",
					Description: "invalid name",
				}

				require.Equal(t, expected, err)
			},
		},
		{
			name: "error with no response body",
			config: &influxdb.HTTPConfig{
				URL:      u,
				Database: "telegraf",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			errFunc: func(t *testing.T, err error) {
				expected := &influxdb.APIError{
					StatusCode: 404,
					Title:      "404 Not Found",
				}

				require.Equal(t, expected, err)
			},
		},
		{
			name: "ok with no response body",
			config: &influxdb.HTTPConfig{
				URL:      u,
				Database: "telegraf",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/query":
					tt.queryHandlerFunc(t, w, r)
					return
				default:
					w.WriteHeader(http.StatusNotFound)
					return
				}
			})

			ctx := context.Background()

			client, err := influxdb.NewHTTPClient(tt.config)
			require.NoError(t, err)
			err = client.CreateDatabase(ctx)
			if tt.errFunc != nil {
				tt.errFunc(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHTTP_Write(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s", ts.Listener.Addr().String()))
	require.NoError(t, err)

	tests := []struct {
		name             string
		config           *influxdb.HTTPConfig
		queryHandlerFunc func(t *testing.T, w http.ResponseWriter, r *http.Request)
		errFunc          func(t *testing.T, err error)
		logFunc          func(t *testing.T, str string)
	}{
		{
			name: "success",
			config: &influxdb.HTTPConfig{
				URL:      u,
				Database: "telegraf",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, r.FormValue("db"), "telegraf")
				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), "cpu value=42")
				w.WriteHeader(http.StatusNoContent)
			},
		},
		{
			name: "send basic auth",
			config: &influxdb.HTTPConfig{
				URL:      u,
				Database: "telegraf",
				Username: "guy",
				Password: "smiley",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				username, password, ok := r.BasicAuth()
				require.True(t, ok)
				require.Equal(t, "guy", username)
				require.Equal(t, "smiley", password)
				w.WriteHeader(http.StatusNoContent)
			},
		},
		{
			name: "send user agent",
			config: &influxdb.HTTPConfig{
				URL:       u,
				Database:  "telegraf",
				UserAgent: "telegraf",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, r.Header.Get("User-Agent"), "telegraf")
				w.WriteHeader(http.StatusNoContent)
			},
		},
		{
			name: "default database",
			config: &influxdb.HTTPConfig{
				URL: u,
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "telegraf", r.FormValue("db"))
				w.WriteHeader(http.StatusNoContent)
			},
		},
		{
			name: "send headers",
			config: &influxdb.HTTPConfig{
				URL: u,
				Headers: map[string]string{
					"A": "B",
					"C": "D",
				},
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, r.Header.Get("A"), "B")
				require.Equal(t, r.Header.Get("C"), "D")
				w.WriteHeader(http.StatusNoContent)
			},
		},
		{
			name: "send retention policy",
			config: &influxdb.HTTPConfig{
				URL:             u,
				Database:        "telegraf",
				RetentionPolicy: "foo",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "foo", r.FormValue("rp"))
				w.WriteHeader(http.StatusNoContent)
			},
		},
		{
			name: "send consistency",
			config: &influxdb.HTTPConfig{
				URL:         u,
				Database:    "telegraf",
				Consistency: "all",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "all", r.FormValue("consistency"))
				w.WriteHeader(http.StatusNoContent)
			},
		},
		{
			name: "hinted handoff not empty no log no error",
			config: &influxdb.HTTPConfig{
				URL:      u,
				Database: "telegraf",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "write failed: hinted handoff queue not empty"}`))
			},
			logFunc: func(t *testing.T, str string) {
				require.False(t, strings.Contains(str, "hinted handoff queue not empty"))
			},
		},
		{
			name: "partial write errors are logged no error",
			config: &influxdb.HTTPConfig{
				URL:      u,
				Database: "telegraf",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "partial write: field type conflict:"}`))
			},
			logFunc: func(t *testing.T, str string) {
				require.Contains(t, str, "partial write")
			},
		},
		{
			name: "parse errors are logged no error",
			config: &influxdb.HTTPConfig{
				URL:      u,
				Database: "telegraf",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": "unable to parse 'cpu value': invalid field format"}`))
			},
			logFunc: func(t *testing.T, str string) {
				require.Contains(t, str, "unable to parse")
			},
		},
		{
			name: "http error",
			config: &influxdb.HTTPConfig{
				URL:      u,
				Database: "telegraf",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			},
			errFunc: func(t *testing.T, err error) {
				expected := &influxdb.APIError{
					StatusCode: 502,
					Title:      "502 Bad Gateway",
				}
				require.Equal(t, expected, err)
			},
		},
		{
			name: "http error with desc",
			config: &influxdb.HTTPConfig{
				URL:      u,
				Database: "telegraf",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"error": "unknown error"}`))
			},
			errFunc: func(t *testing.T, err error) {
				expected := &influxdb.APIError{
					StatusCode:  503,
					Title:       "503 Service Unavailable",
					Description: "unknown error",
				}
				require.Equal(t, expected, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/write":
					tt.queryHandlerFunc(t, w, r)
					return
				default:
					w.WriteHeader(http.StatusNotFound)
					return
				}
			})

			var b bytes.Buffer
			if tt.logFunc != nil {
				log.SetOutput(&b)
			}

			ctx := context.Background()

			m, err := metric.New(
				"cpu",
				map[string]string{},
				map[string]interface{}{
					"value": 42.0,
				},
				time.Unix(0, 0),
			)
			require.NoError(t, err)
			metrics := []telegraf.Metric{m}

			client, err := influxdb.NewHTTPClient(tt.config)
			require.NoError(t, err)
			err = client.Write(ctx, metrics)
			if tt.errFunc != nil {
				tt.errFunc(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.logFunc != nil {
				tt.logFunc(t, b.String())
			}
		})
	}
}

func TestHTTP_WritePathPrefix(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/x/y/z/query":
				w.WriteHeader(http.StatusOK)
				return
			case "/x/y/z/write":
				w.WriteHeader(http.StatusNoContent)
				return
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		},
		),
	)
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s/x/y/z", ts.Listener.Addr().String()))
	require.NoError(t, err)

	ctx := context.Background()

	m, err := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	require.NoError(t, err)
	metrics := []telegraf.Metric{m}

	config := &influxdb.HTTPConfig{
		URL:      u,
		Database: "telegraf",
	}

	client, err := influxdb.NewHTTPClient(config)
	require.NoError(t, err)
	err = client.CreateDatabase(ctx)
	require.NoError(t, err)
	err = client.Write(ctx, metrics)
	require.NoError(t, err)
}

func TestHTTP_WriteContentEncodingGzip(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/write":
				require.Equal(t, r.Header.Get("Content-Encoding"), "gzip")

				gr, err := gzip.NewReader(r.Body)
				require.NoError(t, err)
				body, err := ioutil.ReadAll(gr)
				require.NoError(t, err)

				require.Contains(t, string(body), "cpu value=42")
				w.WriteHeader(http.StatusNoContent)
				return
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		},
		),
	)
	defer ts.Close()

	u, err := url.Parse(fmt.Sprintf("http://%s/", ts.Listener.Addr().String()))
	require.NoError(t, err)

	ctx := context.Background()

	m, err := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
	require.NoError(t, err)
	metrics := []telegraf.Metric{m}

	config := &influxdb.HTTPConfig{
		URL:             u,
		Database:        "telegraf",
		ContentEncoding: "gzip",
	}

	client, err := influxdb.NewHTTPClient(config)
	require.NoError(t, err)
	err = client.Write(ctx, metrics)
	require.NoError(t, err)
}

func TestHTTP_UnixSocket(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf-test")
	if err != nil {
		require.NoError(t, err)
	}
	defer os.RemoveAll(tmpdir)

	sock := path.Join(tmpdir, "test.sock")
	listener, err := net.Listen("unix", sock)
	require.NoError(t, err)

	ts := httptest.NewUnstartedServer(http.NotFoundHandler())
	ts.Listener = listener
	ts.Start()
	defer ts.Close()

	successResponse := []byte(`{"results": [{"statement_id": 0}]}`)

	tests := []struct {
		name             string
		config           *influxdb.HTTPConfig
		database         string
		queryHandlerFunc func(t *testing.T, w http.ResponseWriter, r *http.Request)
		writeHandlerFunc func(t *testing.T, w http.ResponseWriter, r *http.Request)
		errFunc          func(t *testing.T, err error)
	}{
		{
			name: "success",
			config: &influxdb.HTTPConfig{
				URL:      &url.URL{Scheme: "unix", Path: sock},
				Database: "xyzzy",
			},
			queryHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, `CREATE DATABASE "xyzzy"`, r.FormValue("q"))
				w.WriteHeader(http.StatusOK)
				w.Write(successResponse)
			},
			writeHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
				w.Write(successResponse)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/query":
					tt.queryHandlerFunc(t, w, r)
					return
				case "/write":
					tt.queryHandlerFunc(t, w, r)
					return
				default:
					w.WriteHeader(http.StatusNotFound)
					return
				}
			})

			ctx := context.Background()

			client, err := influxdb.NewHTTPClient(tt.config)
			require.NoError(t, err)
			err = client.CreateDatabase(ctx)
			if tt.errFunc != nil {
				tt.errFunc(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
