//go:build darwin
// +build darwin

package wireless

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadWirelessTable(t *testing.T) {
	w := Wireless{}
	// line of input
	macInput, err := ioutil.ReadFile("testdata/wireless_mac_test.txt")
	require.NoError(t, err)

	// the map of data we expect.
	macFields := map[string]interface{}{
		"agrCtlRSSI":      int64(-42),
		"agrExtRSSI":      int64(0),
		"agrCtlNoise":     int64(-92),
		"agrExtNoise":     int64(0),
		"lastTxRate":      int64(300),
		"maxRate":         int64(450),
		"lastAssocStatus": int64(0),
		"MCS":             int64(15),
	}
	// the tags we expect
	macTags := map[string]string{
		"state":       "running",
		"op_mode":     "station",
		"802.11_auth": "open",
		"link_auth":   "wpa2-psk",
		"BSSID":       "12:34:56:78:9a:bc",
		"SSID":        "Foo_Bar",
		"channel":     "157,1",
	}

	// load the table from the input.
	gotFields, gotTags := w.loadMacWirelessTable([]byte(macInput))
	require.NoError(t, err)
	require.Equal(t, gotFields, macFields)
	require.Equal(t, gotTags, macTags)
}
