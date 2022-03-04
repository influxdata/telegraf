//go:build linux
// +build linux

package wireless

import (
	"io/ioutil"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func TestLoadWirelessTable(t *testing.T) {
	testInput, err := ioutil.ReadFile("testdata/wireless_linux_test.txt")
	require.NoError(t, err)
	expectedMetrics := []*wirelessInterface{
		{
			Interface: "wlan0",
			Status:    int64(0000),
			Link:      int64(60),
			Level:     int64(-50),
			Noise:     int64(-256),
			Nwid:      int64(0),
			Crypt:     int64(0),
			Frag:      int64(0),
			Retry:     int64(1525),
			Misc:      int64(0),
			Beacon:    int64(0),
		},
		{
			Interface: "wlan1",
			Status:    int64(0000),
			Link:      int64(70),
			Level:     int64(-39),
			Noise:     int64(-256),
			Nwid:      int64(0),
			Crypt:     int64(0),
			Frag:      int64(0),
			Retry:     int64(12096),
			Misc:      int64(191188),
			Beacon:    int64(0),
		},
	}

	w := Wireless{
		Log: testutil.Logger{},
	}
	metrics, err := w.loadWirelessTable(testInput)
	require.NoError(t, err)

	as := require.New(t)
	as.Equal(metrics, expectedMetrics)
}
