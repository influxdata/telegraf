package spark

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestSparkMeasurements(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	a := &Spark{
		YarnServer: testutil.GetLocalHost() + ":8088",
	}

	var acc testutil.Accumulator

	err := a.Gather(&acc)
	require.NoError(t, err)
}
