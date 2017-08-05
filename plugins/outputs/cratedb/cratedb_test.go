package cratedb

import (
	"os"
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

	c := &CrateDB{
		URL:         testURL(),
		Table:       "test",
		Timeout:     internal.Duration{Duration: time.Second * 5},
		TableCreate: true,
	}

	require.NoError(t, c.Connect())
	require.NoError(t, c.Write(testutil.MockMetrics()))
	require.NoError(t, c.Close())

}

func testURL() string {
	url := os.Getenv("CRATE_URL")
	if url == "" {
		// @TODO use telegraf helper func for hostname
		return "postgres://localhost:6543/test?sslmode=disable"
	}
	return url
}
