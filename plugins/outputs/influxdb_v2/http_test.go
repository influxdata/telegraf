package influxdb_v2_test

import (
	"context"
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
