package influxdb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

func TestUDPInflux(t *testing.T) {
	i := InfluxDB{
		URLs: []string{"udp://localhost:8089"},
	}

	err := i.Connect()
	require.NoError(t, err)
	err = i.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestHTTPInflux(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"results":[{}]}`)
	}))
	defer ts.Close()

	i := InfluxDB{
		URLs: []string{ts.URL},
	}

	err := i.Connect()
	require.NoError(t, err)
	err = i.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestDownsampling_mean(t *testing.T) {
	ds := &Downsampling{}
	metricA, err := telegraf.NewMetric(
		"earthshaker",
		map[string]string{},
		map[string]interface{}{
			"damage":       "high",
			"agility":      12,
			"strength":     120,
			"intelligence": 60,
		},
		time.Now(),
	)

	metricB, err := telegraf.NewMetric(
		"sven",
		map[string]string{},
		map[string]interface{}{
			"strength":     80,
			"intelligence": 140,
		},
		time.Now(),
	)
	require.NoError(t, err)

	err = ds.Add(metricA)
	require.NoError(t, err)

	err = ds.Add(metricB)
	require.NoError(t, err)

	aggr, err := ds.Mean("strength", "intelligence", "power")
	require.NoError(t, err)

	require.Equal(t, int64(100), aggr.Fields()["strength"])
	require.Equal(t, int64(100), aggr.Fields()["intelligence"])
	require.Equal(t, int64(0), aggr.Fields()["power"])
}
