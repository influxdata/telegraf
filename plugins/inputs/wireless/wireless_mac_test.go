//go:build darwin
// +build darwin

package wireless

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	field_len = 11
	tag_len   = 5
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
		"BSSID":           "",
		"SSID":            "Foo_Bar",
		"channel":         "157,1",
	}
	// the tags we expect
	macTags := map[string]string{
		"state":       "running",
		"op_mode":     "station",
		"802.11_auth": "open",
		"link_auth":   "wpa2-psk",
		"interface":   "airport",
	}

	// load the table from the input.
	gotFields, gotTags, err := w.loadMacWirelessTable([]byte(macInput))
	require.NoError(t, err)
	require.Equal(t, len(gotFields), field_len, "got %d fields, expected %d", len(gotFields), field_len)
	require.Equal(t, gotFields, macFields, "want %+v, got %+v", macFields, gotFields)
	require.Equal(t, len(gotTags), tag_len, "got %d tags, expected %d", len(gotTags), tag_len)
	require.Equal(t, gotTags, macTags, "want %+v, got %+v", macTags, gotTags)
}
