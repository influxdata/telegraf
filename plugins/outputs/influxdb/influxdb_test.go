package influxdb

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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

	err := ds.Add(testutil.TestMetric(120))
	require.NoError(t, err)

	err = ds.Add(testutil.TestMetric(80))
	require.NoError(t, err)

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

func TestDownsamling_sum(t *testing.T) {
	ds := &Downsampling{}

	err := ds.Add(testutil.TestMetric(120))
	require.NoError(t, err)

	err = ds.Add(testutil.TestMetric(80))
	require.NoError(t, err)

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
	ds := &Downsampling{}

	err := ds.Add(testutil.TestMetric(120))
	require.NoError(t, err)

	err = ds.Add(testutil.TestMetric(80))
	require.NoError(t, err)

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

	ds.Aggregations = make(map[string][]Aggregation)
	ds.AddAggregations(aggregations...)

	aggr, err := ds.Aggregate()
	require.NoError(t, err)

	require.Equal(t, int64(100), aggr.Fields()["mean_value"])
	require.Equal(t, int64(200), aggr.Fields()["sum_value"])

}
