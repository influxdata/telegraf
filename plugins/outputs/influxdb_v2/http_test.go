package influxdb_v2_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	influxdb "github.com/influxdata/telegraf/plugins/outputs/influxdb_v2"
	"github.com/influxdata/telegraf/testutil"
)

func genURL(u string) *url.URL {
	address, _ := url.Parse(u)
	return address
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
				err := r.ParseForm()
				require.NoError(t, err)
				require.Equal(t, r.Form["bucket"], []string{"foo"})

				body, err := io.ReadAll(r.Body)
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

func TestTooLargeWriteRetry(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/api/v2/write":
				err := r.ParseForm()
				require.NoError(t, err)

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				// Ensure metric body size is small
				if len(body) > 16 {
					w.WriteHeader(http.StatusRequestEntityTooLarge)
				} else {
					w.WriteHeader(http.StatusNoContent)
				}

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
		Log:              testutil.Logger{},
	}

	client, err := influxdb.NewHTTPClient(config)
	require.NoError(t, err)

	// Together the metric batch size is too big, split up, we get success
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
		testutil.MustMetric(
			"cpu",
			map[string]string{
				"bucket": "bar",
			},
			map[string]interface{}{
				"value": 99.0,
			},
			time.Unix(0, 0),
		),
	}

	ctx := context.Background()
	err = client.Write(ctx, metrics)
	require.NoError(t, err)

	// These metrics are too big, even after splitting in half, expect error
	hugeMetrics := []telegraf.Metric{
		testutil.MustMetric(
			"reallyLargeMetric",
			map[string]string{
				"bucket": "foobar",
			},
			map[string]interface{}{
				"value": 123.456,
			},
			time.Unix(0, 0),
		),
		testutil.MustMetric(
			"evenBiggerMetric",
			map[string]string{
				"bucket": "fizzbuzzbang",
			},
			map[string]interface{}{
				"value": 999.999,
			},
			time.Unix(0, 0),
		),
	}

	err = client.Write(ctx, hugeMetrics)
	require.Error(t, err)
}
