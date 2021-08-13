package internetSpeedMonitor

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInternetConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}
	speedMonitor := &SpeedMonitor{
		EnableFileDownload: false,
		Measurement:        "internet_speed",
	}

	assert.Equal(t, speedMonitor.Measurement, "internet_speed")
	assert.Equal(t, speedMonitor.EnableFileDownload, false)
	assert.Equal(t, speedMonitor.Description(), "Monitors internet speed in the network")

}

func TestGathering(t *testing.T) {
	speedMonitor := &SpeedMonitor{
		EnableFileDownload: false,
		Measurement:        "internet_speed",
		Log:                testutil.Logger{},
	}

	acc := &testutil.Accumulator{}

	require.NoError(t, speedMonitor.Gather(acc))
}

func TestDataGen(t *testing.T) {
	speedMonitor := &SpeedMonitor{
		EnableFileDownload: false,
		Measurement:        "internet_speed",
		Log:                testutil.Logger{},
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, speedMonitor.Gather(acc))

	metric, ok := acc.Get("internet_speed")
	require.True(t, ok)

	tags := metric.Tags

	fields := metric.Fields

	acc.AssertContainsTaggedFields(t, "internet_speed", fields, tags)
}
