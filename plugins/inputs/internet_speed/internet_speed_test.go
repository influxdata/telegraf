package internet_speed

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestGathering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	internetSpeed := &InternetSpeed{
		EnableFileDownload: false,
		Log:                testutil.Logger{},
	}

	acc := &testutil.Accumulator{}

	require.NoError(t, internetSpeed.Gather(acc))
}

func TestDataGen(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	internetSpeed := &InternetSpeed{
		EnableFileDownload: false,
		Log:                testutil.Logger{},
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, internetSpeed.Gather(acc))

	metric, ok := acc.Get("internet_speed")
	require.True(t, ok)

	tags := metric.Tags

	fields := metric.Fields

	acc.AssertContainsTaggedFields(t, "internet_speed", fields, tags)
}
