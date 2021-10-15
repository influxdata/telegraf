package mongodb

import (
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/metric"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func getMetrics() []telegraf.Metric {
	m := metric.New(
		"cpu",
		map[string]string{},
		map[string]interface{}{
			"value": 42.0,
		},
		time.Unix(0, 0),
	)
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
