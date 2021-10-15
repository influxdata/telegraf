package mongodb

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	"testing"
)

func getMetrics() []telegraf.Metric {
	m := testutil.TestMetric("mymetric")
	metrics := []telegraf.Metric{m}
	return metrics
}

func TestConnectNoAuthAndInsertDocument(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	s := MongoDB{
		Dsn:                "mongodb://localhost:27017",
		AuthenticationType: "NONE",
		MetricDatabase:     "myMetricDatabase",
		MetricGranularity:  "minutes",
		AllowTLSInsecure:   false,
		TTL:                "15d",
	}

	// connect to mongodb
	err := s.Connect()
	require.NoError(t, err)

	// create time series collection when it doesn't exist
	myTestMetricName := "testMetricName"
	if !s.DoesCollectionExist(myTestMetricName) {
		err = s.MongoDBCreateTimeSeriesCollection(myTestMetricName)
		require.NoError(t, err)
	}

	// test insert
	err = s.Write(getMetrics())
	require.NoError(t, err)

	// cleanup
	err = s.Close()
	require.NoError(t, err)
}
