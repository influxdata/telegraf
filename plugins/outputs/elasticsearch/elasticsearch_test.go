package elasticsearch

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	serverhost := "http://" + testutil.GetLocalHost() + ":9200"
	e := &Elasticsearch{
		ServerHost: serverhost,
		IndexName:  "littletest",
	}

	// Verify that we can connect to the ElasticSearch
	err := e.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the ElasticSearch
	err = e.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
