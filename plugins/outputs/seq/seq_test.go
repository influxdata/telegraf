package seq

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestConnectAndWrite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	url := testutil.GetLocalHost() + ":5341"

	r := &Seq{
		SeqInstance: url,
	}

	err := r.Connect()
	require.NoError(t, err)

	err = r.Write(testutil.MockMetrics())
	require.NoError(t, err)
}
