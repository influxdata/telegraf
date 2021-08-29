//go:build linux
// +build linux

package wireless

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testInput = []byte(`Inter-| sta-|   Quality        |   Discarded packets               | Missed | WE
 face | tus | link level noise |  nwid  crypt   frag  retry   misc | beacon | 22
 wlan0: 0000   60.  -50.  -256        0      0      0   1525      0        0
 wlan1: 0000   70.  -39.  -256        0      0      0  12096 191188        0`)

func TestLoadWirelessTable(t *testing.T) {
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
	metrics, err := loadWirelessTable(testInput)
	if err != nil {
		t.Fatal(err)
	}

	as := assert.New(t)
	as.Equal(metrics, expectedMetrics)
}
