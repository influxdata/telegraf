package elasticsearch

import (
	"testing"
	"time"

	"github.com/influxdata/telegraf/internal"
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

	// Verify that we can connect to the ElasticSearch
	err := e.Connect()
	require.NoError(t, err)

	// Verify that we can successfully write data to the ElasticSearch
	err = e.Write(testutil.MockMetrics())
	require.NoError(t, err)

}
