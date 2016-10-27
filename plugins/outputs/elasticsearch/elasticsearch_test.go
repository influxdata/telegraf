package elasticsearch

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/influxdata/telegraf"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	serverhost := "http://" + testutil.GetLocalHost() + ":19200"

	e := &Elasticsearch{
		ServerHost:       serverhost,
		IndexName:        "littletest%Y%m%d",
		NumberOfShards:   2,
		NumberOfReplicas: 2,
	}

	// Verify that we can connect to the ElasticSearch
	err := e.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the ElasticSearch
	err = e.Write(testutil.MockMetrics())
	require.NoError(t, err)

	// Verify if metric sent has same data on Elasticsearch
	metrictest, _ := telegraf.NewMetric(
		"my_measurement",
		map[string]string{"host": "192.168.0.1"},
		map[string]interface{}{"value": float64(3.14)},
		time.Date(2010, time.November, 10, 23, 0, 0, 0, time.UTC),
	)

	messageid, err := e.WriteOneMessage(metrictest)
	require.NoError(t, err)

	get1, errGet := e.Client.Get().
		Index(e.IndexName).
		Type(metrictest.Name()).
		Id(messageid).
		Do()
	require.NoError(t, errGet)

	require.Equal(t, true, get1.Found, "Message not found on Elasticsearch.")

	require.NotEqual(t, nil, get1.Source, "Source not found on Elasticsearch.")

	var dat map[string]interface{}
	err = json.Unmarshal(*get1.Source, &dat)
	require.NoError(t, err)

	require.Equal(t, "192.168.0.1", dat["host"], "Values of Host are not the same.")
	require.Equal(t, "2010-11-10T23:00:00Z", dat["created"], "Values of Created are not the same.")
	require.Equal(t, 3.14, dat["value"], "Values of Value are not the same.")

}
