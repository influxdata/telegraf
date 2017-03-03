package elasticsearch

import (
	"math"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	urls := []string{"http://" + testutil.GetLocalHost() + ":9200"}

	e := &Elasticsearch{
		URLs:                urls,
		IndexName:           "littletest-%Y.%m.%d",
		Timeout:             internal.Duration{Duration: time.Second * 5},
		ManageTemplate:      true,
		TemplateName:        "telegraf",
		OverwriteTemplate:   true,
		HealthCheckInterval: internal.Duration{Duration: time.Second * 10},
	}

	// Verify that we can connect to Elasticsearch
	err := e.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to Elasticsearch
	err = e.Write(testutil.MockMetrics())
	require.NoError(t, err)

}

func TestBigValue(t *testing.T) {
	urls := []string{"http://" + testutil.GetLocalHost() + ":9200"}

	e := &Elasticsearch{
		URLs:                urls,
		IndexName:           "littletest-%Y.%m.%d",
		Timeout:             internal.Duration{Duration: time.Second * 5},
		ManageTemplate:      true,
		TemplateName:        "telegraf",
		OverwriteTemplate:   true,
		HealthCheckInterval: internal.Duration{Duration: time.Second * 10},
	}

	// Init metrics
	m1, _ := metric.New(
		"mymeasurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{
			"myvalue1": float64(-9223372036854776000),
			"myvalue2": float64(math.MaxUint64 * -10),
			"myvalue3": float64(math.MaxUint64 * 10),
			"myvalue4": float64(math.MaxFloat64),
			"myvalue5": float64(0.000000000000000000000000000000000000000000000001),
			"myvalue6": float64(-0.000000000000000000000000000000000000000000000001),
		},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)
	// Prepare point list
	var metrics []telegraf.Metric
	metrics = append(metrics, m1)

	// Verify that we can connect to Elasticsearch
	err := e.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to Elasticsearch
	err = e.Write(metrics)
	require.NoError(t, err)

}
