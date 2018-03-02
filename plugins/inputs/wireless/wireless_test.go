// +build linux

package wireless

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testInput = []byte(`Inter-| sta-|   Quality        |   Discarded packets               | Missed | WE
 face | tus | link level noise |  nwid  crypt   frag  retry   misc | beacon | 22
 wlan0: 0000   60.  -50.  -256        0      0      0   1525      0        0`)

func TestLoadWirelessTable(t *testing.T) {
	expectedMetrics := map[string]interface{}{
		"status":        int64(0),
		"link":          int64(60),
		"level":         int64(-50),
		"noise":         int64(-256),
		"nwid":          int64(0),
		"crypt":         int64(0),
		"frag":          int64(0),
		"retry":         int64(1525),
		"misc":          int64(0),
		"missed_beacon": int64(0),
	}
	expectedTags := map[string]string{
		"interface": "wlan0",
	}

	metrics, tags, err := loadWirelessTable(testInput, false)
	if err != nil {
		t.Fatal(err)
	}

	as := assert.New(t)
	for k := range metrics {
		as.Equal(metrics[k], expectedMetrics[k])
	}
	for k := range tags {
		as.Equal(tags[k], expectedTags[k])
	}
}
