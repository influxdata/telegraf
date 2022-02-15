//go:build darwin
// +build darwin

package wireless

import (
	"runtime"
	"testing"
)

func TestLoadWirelessTable(t *testing.T) {
	goOS := runtime.GOOS
	ns := Wireless{}
	ns.DumpZeros = true
	if goOS == "darwin" {
		// line of input
		macInput := `agrCtlRSSI: -42
     agrExtRSSI: 0
    agrCtlNoise: -92
    agrExtNoise: 0
          state: running
        op mode: station
     lastTxRate: 300
        maxRate: 450
lastAssocStatus: 0
    802.11 auth: open
      link auth: wpa2-psk
          BSSID:
           SSID: Foo Bar
            MCS: 15
        channel: 157,1`
		// the headers we expect from that line of input

		// the map of data we expect.
		macParsed := map[string]interface{}{
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
			"BSSID":       "",
			"SSID":        "Foo_Bar",
			"interface":   "airport",
		}

		// load the table from the input.
		gotPoints, gotTags, err := ns.loadMacWirelessTable([]byte(macInput), true)
		if err != nil {
			t.Fatal(err)
		}
		if len(gotPoints) == 0 {
			t.Fatalf("want %+v, got %+v", macParsed, gotPoints)
		}
		for key := range macParsed {
			if macParsed[key].(int64) != gotPoints[key].(int64) {
				t.Fatalf("want %+v, got %+v", macParsed[key], gotPoints[key])
			}
		}
		for key := range macTags {
			if macTags[key] != gotTags[key] {
				t.Fatalf("want %+v, got %+v", macTags[key], gotTags[key])
			}

		}
	} else {
		t.Fatalf("unsupported OS %s", goOS)
	}
}
