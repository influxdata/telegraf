package internet_speed

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/testutil"
)

func TestGathering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	internetSpeed := &InternetSpeed{
		MemorySavingMode: true,
		Log:              testutil.Logger{},
	}
	require.NoError(t, internetSpeed.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, internetSpeed.Gather(acc))
}

func TestDataGen(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	internetSpeed := &InternetSpeed{
		MemorySavingMode: true,
		Log:              testutil.Logger{},
	}
	require.NoError(t, internetSpeed.Init())

	acc := &testutil.Accumulator{}
	require.NoError(t, internetSpeed.Gather(acc))

	metric, ok := acc.Get("internet_speed")
	require.True(t, ok)
	acc.AssertContainsTaggedFields(t, "internet_speed", metric.Fields, metric.Tags)
}
