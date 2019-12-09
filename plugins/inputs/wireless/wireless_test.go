package wireless

import (
	"fmt"
	"runtime"
	"testing"
)

func TestLoadWirelessTable(t *testing.T) {
	goOS := runtime.GOOS
	if goOS == "linux" {
		// line of input
		linuxInput := `Inter-| sta-|   Quality        |   Discarded packets               | Missed | WE
 face | tus | link level noise |  nwid  crypt   frag  retry   misc | beacon | 22
 wlan0: 0000   96.  -65.  -256.       0      0      0      0      0        0`
		// the headers we expect from that line of input
		linuxHeaders := []string{"status", "quality_link", "quality_level", "quality_noise", "discarded_packets_nwid", "discarded_packets_crypt",
			"discarded_packets_frag", "discarded_packets_retry", "discarded_packets_misc", "missed_beacon", "we_22"}
		// the map of data we expect.
		linuxParsed := map[string]interface{}{
			"status":                  int64(0),
			"quality_link":            int64(96),
			"quality_level":           int64(-65),
			"quality_noise":           int64(-256),
			"discarded_packets_nwid":  int64(0),
			"discarded_packets_crypt": int64(0),
			"discarded_packets_frag":  int64(0),
			"discarded_packets_retry": int64(0),
			"discarded_packets_misc":  int64(0),
			"missed_beacon":           int64(0),
			"we_22":                   int64(0),
		}
		// the tags we expect
		linuxTestTags := map[string]interface{}{
			"interface": "wlan0:",
		}
		// Map of the entries we get back from the table
		linuxEntries := map[string]interface{}{}

		// load the table from the input.
		got, err := loadLinuxWirelessTable([]byte(linuxInput), true)
		if err != nil {
			t.Fatal(err)
		}
		// the WirelessData struct holds arrays of the values, so
		// move them into appropriate maps.
		tags := map[string]string{}
		for x := 0; x < len(got.Tags); x++ {
			entries := map[string]interface{}{}
			fmt.Println("Tag: ", got.Tags[x])
			tags = map[string]string{
				"interface": got.Tags[x],
			}
			for z := 0; z < len(got.Data[x]); z++ {
				linuxEntries[got.Headers[z]] = got.Data[x][z]
				fmt.Println(entries[got.Headers[z]])
			}
		}
		// make sure we got the same number of headers back we expect.
		if len(got.Headers) != len(linuxHeaders) {
			t.Fatalf("want %+v, got %+v", linuxHeaders, got.Headers)
		}
		// create the data map
		for z := 0; z < len(got.Data[0]); z++ {
			linuxEntries[got.Headers[z]] = got.Data[0][z]
		}
		// verify the data map
		for key := range linuxParsed {
			if linuxParsed[key] != linuxEntries[key] {
				t.Fatalf("want %+v, got %+v", linuxParsed[key], linuxEntries[key])
			}
		}
		// verify the tag map
		for key := range tags {
			if linuxTestTags[key] != tags[key] {
				t.Fatalf("want %+v, got %+v", linuxTestTags[key], tags[key])
			}
		}
	} else if goOS == "darwin" {
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
          BSSID: 5c:99:99:99:9:99
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

		macTags := map[string]string{
			"state":       "running",
			"op_mode":     "station",
			"802.11_auth": "open",
			"link_auth":   "wpa2-psk",
			"BSSID":       "5c:99:99:99:9:99",
			"SSID":        "Foo_Bar",
			"interface":   "airport",
		}

		// load the table from the input.
		got, gotTags, err := loadMacWirelessTable([]byte(macInput), false)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) == 0 {
			t.Fatalf("want %+v, got %+v", macParsed, got)
		}
		for key := range macParsed {
			if macParsed[key].(int64) != got[key].(int64) {
				t.Fatalf("want %+v, got %+v", macParsed[key], got[key])
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
