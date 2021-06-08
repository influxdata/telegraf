package influxdb_v2_test

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	influxdb "github.com/influxdata/telegraf/plugins/outputs/influxdb_v2"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type createBucketRequest struct {
	Name           string          `json:"name"`
	OrgID          string          `json:"orgID"`
	RetentionRules []retentionRule `json:"retentionRules"`
}

type retentionRule struct {
	EverySeconds int64  `json:"everySeconds"`
	Type         string `json:"type"`
}

func genURL(u string) *url.URL {
	URL, _ := url.Parse(u)
	return URL
}
func TestNewHTTPClient(t *testing.T) {
	tests := []struct {
		err bool
		cfg *influxdb.HTTPConfig
	}{
		{
			err: true,
			cfg: &influxdb.HTTPConfig{},
		},
		{
			err: true,
			cfg: &influxdb.HTTPConfig{
				URL: genURL("udp://localhost:9999"),
			},
		},
		{
			cfg: &influxdb.HTTPConfig{
				URL: genURL("unix://var/run/influxd.sock"),
			},
		},
	}

	for i := range tests {
		client, err := influxdb.NewHTTPClient(tests[i].cfg)
		if !tests[i].err {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			t.Log(err)
		}
		if err == nil {
			client.URL()
		}
	}
}

func TestWriteBucketTagWorksOnRetry(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v2/write":
				r.ParseForm()
				require.Equal(t, r.Form["bucket"], []string{"foo"})

				body, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), "cpu value=42")

				w.WriteHeader(http.StatusNoContent)
				return
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		}),
	)
	defer ts.Close()

	addr := &url.URL{
		Scheme: "http",
		Host:   ts.Listener.Addr().String(),
	}

	config := &influxdb.HTTPConfig{
		URL:              addr,
		Bucket:           "telegraf",
		BucketTag:        "bucket",
		ExcludeBucketTag: true,
	}

	client, err := influxdb.NewHTTPClient(config)
	require.NoError(t, err)

	metrics := []telegraf.Metric{
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"bucket": "foo",
			},
			map[string]interface{}{
				"value": 42.0,
			},
			time.Unix(0, 0),
		),
	}

	ctx := context.Background()
	err = client.Write(ctx, metrics)
	require.NoError(t, err)
	err = client.Write(ctx, metrics)
	require.NoError(t, err)
}

func parseBody(t *testing.T, reader io.ReadCloser) createBucketRequest {
	var body createBucketRequest
	require.NoError(t, json.NewDecoder(reader).Decode(&body))
	return body
}

func TestCreateBucket(t *testing.T) {
	ts := httptest.NewServer(http.NotFoundHandler())
	defer ts.Close()

	u, err := url.Parse("http://" + ts.Listener.Addr().String())
	require.NoError(t, err)

	tests := []struct {
		name              string
		config            influxdb.HTTPConfig
		err               bool
		createHandlerFunc func(*testing.T, http.ResponseWriter, *http.Request)
		orgIDHandlerFunc  func(*testing.T, http.ResponseWriter, *http.Request)
	}{
		{
			name:   "success",
			config: influxdb.HTTPConfig{URL: u, Bucket: "bucket", Organization: "telegraf"},
			createHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				body := parseBody(t, r.Body)
				require.Equal(t, "bucket", body.Name)
				require.Equal(t, "0123456789abcdef", body.OrgID)
				require.Len(t, body.RetentionRules, 1)
				require.Equal(t, "expire", body.RetentionRules[0].Type)
				require.Equal(t, int64(0), body.RetentionRules[0].EverySeconds)
				w.WriteHeader(http.StatusCreated)
			},
			orgIDHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "telegraf", r.URL.Query().Get("org"))
				require.Equal(t, "1", r.URL.Query().Get("limit"))
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"orgs": [{"id": "0123456789abcdef"}]}`))
				require.NoError(t, err)
			},
		},
		{
			name:   "custom retention rule",
			config: influxdb.HTTPConfig{URL: u, Bucket: "bucket", Organization: "telegraf", DefaultBucketRetention: 120},
			createHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				body := parseBody(t, r.Body)
				require.Len(t, body.RetentionRules, 1)
				require.Equal(t, "expire", body.RetentionRules[0].Type)
				require.Equal(t, int64(120), body.RetentionRules[0].EverySeconds)
				w.WriteHeader(http.StatusCreated)
			},
			orgIDHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"orgs": [{"id": "0123456789abcdef"}]}`))
				require.NoError(t, err)
			},
		},
		{
			name:   "bad permissions",
			err:    true,
			config: influxdb.HTTPConfig{URL: u, Bucket: "bucket", Organization: "telegraf"},
			orgIDHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			},
		},
		{
			name:   "non-existent org",
			err:    true,
			config: influxdb.HTTPConfig{URL: u, Bucket: "bucket", Organization: "non-existent"},
			orgIDHandlerFunc: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(`{"orgs": []}`))
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v2/orgs":
					tt.orgIDHandlerFunc(t, w, r)
				case "/api/v2/buckets":
					tt.createHandlerFunc(t, w, r)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			})

			ctx := context.Background()

			client, err := influxdb.NewHTTPClient(&tt.config)
			require.NoError(t, err)

			err = client.CreateBucket(ctx, client.Bucket)
			if tt.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
