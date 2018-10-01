package mac_wireless

import (
	"testing"
)

func TestLoadWirelessTable(t *testing.T) {
	// line of input
	input := `agrCtlRSSI: -42
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
          BSSID: 5c:99:99:99:9:99
           SSID: Foo Bar
            MCS: 15
        channel: 157,1`
	// the headers we expect from that line of input

	// the map of data we expect.
	parsed := map[string]interface{}{
		"agrCtlRSSI":      int64(-42),
		"agrExtRSSI":      int64(0),
		"agrCtlNoise":     int64(-92),
		"agrExtNoise":     int64(0),
		"lastTxRate":      int64(300),
		"maxRate":         int64(450),
		"lastAssocStatus": int64(0),
		"MCS":             int64(15),
	}

	tags := map[string]string{
		"state":       "running",
		"op_mode":     "station",
		"802.11_auth": "open",
		"link_auth":   "wpa2-psk",
		"BSSID":       "5c:99:99:99:9:99",
		"SSID":        "Foo_Bar",
		"interface":   "airport",
	}

	// load the table from the input.
	got, got_tags, err := loadWirelessTable([]byte(input), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 {
		t.Fatalf("want %+v, got %+v", parsed, got)
	}
	for key := range parsed {
		if parsed[key].(int64) != got[key].(int64) {
			t.Fatalf("want %+v, got %+v", parsed[key], got[key])
		}
	}
	for key := range tags {
		if tags[key] != got_tags[key] {
			t.Fatalf("want %+v, got %+v", tags[key], got_tags[key])
		}

	}
}
