//go:build !race && linux
// +build !race,linux

package sysstat

import (
	"os/exec"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

// TestInterval verifies that the correct interval is created. It is not
// run with -race option, because in that scenario interval between the two
// Gather calls is greater than wantedInterval.
func TestInterval(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test with sleep in short mode.")
	}
	// overwriting exec commands with mock commands
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()
	var acc testutil.Accumulator

	s := &Sysstat{
		Log:        testutil.Logger{},
		interval:   0,
		Sadc:       "/usr/lib/sa/sadc",
		Sadf:       "/usr/bin/sadf",
		Group:      false,
		Activities: []string{"DISK", "SNMP"},
		Options: map[string]string{
			"C": "cpu",
			"d": "disk",
		},
		DeviceTags: map[string][]map[string]string{
			"sda": {
				{
					"vg": "rootvg",
				},
			},
		},
	}
	require.NoError(t, s.Init())
	require.NoError(t, acc.GatherError(s.Gather))

	wantedInterval := 3
	time.Sleep(time.Duration(wantedInterval) * time.Second)

	require.NoError(t, acc.GatherError(s.Gather))
	require.Equalf(t, wantedInterval, s.interval, "wrong interval: got %d, want %d", s.interval, wantedInterval)
}
