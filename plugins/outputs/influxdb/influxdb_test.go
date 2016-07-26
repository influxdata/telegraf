package influxdb

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

func TestUDPInflux(t *testing.T) {
	i := InfluxDB{
		URLs:        []string{"udp://localhost:8089"},
		Downsampler: &Downsampling{},
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
		URLs:        []string{ts.URL},
		Downsampler: &Downsampling{},
	}

	err := i.Connect()
	require.NoError(t, err)
	err = i.Write(testutil.MockMetrics())
	require.NoError(t, err)
}

func TestDownsampling_mean(t *testing.T) {
	ds := NewDownsampling("downsampling", time.Minute)
	ds.Add(testutil.TestMetric(120))
	ds.Add(testutil.TestMetric(80))

	aggregations := []Aggregation{
		Aggregation{
			FieldName: "value",
			FuncName:  "mean",
			Alias:     "mean_value",
		},
	}

	aggr, err := ds.Mean(aggregations...)
	require.NoError(t, err)

	require.Equal(t, int64(100), aggr.Fields()["mean_value"])
}

func TestDownsampling_sum(t *testing.T) {
	ds := NewDownsampling("downsampling", time.Minute)
	ds.Add(testutil.TestMetric(120))
	ds.Add(testutil.TestMetric(80))

	aggregations := []Aggregation{
		Aggregation{
			FieldName: "value",
			FuncName:  "mean",
			Alias:     "sum_value",
		},
	}
	aggr, err := ds.Sum(aggregations...)
	require.NoError(t, err)

	require.Equal(t, int64(200), aggr.Fields()["sum_value"])
}

func TestDownsampling_aggregate(t *testing.T) {
	ds := NewDownsampling("downsampling", time.Minute)

	ds.Add(testutil.TestMetric(120))
	ds.Add(testutil.TestMetric(80))

	aggregations := []Aggregation{
		Aggregation{
			FieldName: "value",
			FuncName:  "mean",
			Alias:     "mean_value",
		},
		Aggregation{
			FieldName: "value",
			FuncName:  "sum",
			Alias:     "sum_value",
		},
	}

	ds.AddAggregations(aggregations...)

	aggr, err := ds.Aggregate()
	require.NoError(t, err)

	require.Equal(t, int64(100), aggr.Fields()["mean_value"])
	require.Equal(t, int64(200), aggr.Fields()["sum_value"])

}

func TestDownsampling_run(t *testing.T) {
	var (
		sum, count int32
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var createDatabaseQuery = "CREATE DATABASE IF NOT EXISTS \"\""

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"results":[{}]}`)

		err := r.ParseForm()
		require.NoError(t, err)

		q := r.Form.Get("q")
		if q == createDatabaseQuery {
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)

		points, err := models.ParsePoints(body)
		require.NoError(t, err)

		if len(points) == 0 {
			return
		}

		mean, ok := points[0].Fields()["mean_value"]
		if !ok {
			return
		}

		want := atomic.LoadInt32(&sum) / atomic.LoadInt32(&count)
		atomic.StoreInt32(&sum, 0)
		atomic.StoreInt32(&count, 0)

		require.EqualValues(t, want, mean)

	}))
	defer ts.Close()

	downsampler := &Downsampling{
		TimeRange: time.Duration(time.Second * 10),
		Name:      "downsampling",
	}

	downsampler.Aggregations = make(map[string][]Aggregation)
	downsampler.AddAggregations(Aggregation{
		FieldName: "value",
		FuncName:  "mean",
		Alias:     "mean_value",
	})

	influxdb := &InfluxDB{
		Downsampler: downsampler,
		URLs:        []string{ts.URL},
	}
	go influxdb.Run()

	rand.Seed(time.Now().Unix())

	tick := time.Tick(3 * time.Second)
	after := time.After(12 * time.Second)

	for {
		select {
		case <-tick:
			atomic.AddInt32(&count, 1)
			val := rand.Int31n(120)
			atomic.AddInt32(&sum, val)
			err := influxdb.Write([]telegraf.Metric{testutil.TestMetric(val)})
			require.NoError(t, err)
		case <-after:
			return
		}
	}
}
