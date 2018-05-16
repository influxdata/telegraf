package linux_wireless

import (
	"fmt"
	"testing"
)

func TestLoadWirelessTable(t *testing.T) {
	// line of input
	input := `Inter-| sta-|   Quality        |   Discarded packets               | Missed | WE
 face | tus | link level noise |  nwid  crypt   frag  retry   misc | beacon | 22
 wlan0: 0000   96.  -65.  -256.       0      0      0      0      0        0`
	// the headers we expect from that line of input
	headers := []string{"status", "quality_link", "quality_level", "quality_noise", "discarded_packets_nwid", "discarded_packets_crypt",
		"discarded_packets_frag", "discarded_packets_retry", "discarded_packets_misc", "missed_beacon", "we_22"}
	// the map of data we expect.
	parsed := map[string]interface{}{
		"status":                  int64(0),
		"quality_link":            int64(0),
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
	test_tags := map[string]interface{}{
		"interface": "wlan0:",
	}
	// Map of the entries we get back from the table
	entries := map[string]interface{}{}

	// load the table from the input.
	got, err := loadWirelessTable([]byte(input), true)
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
			entries[got.Headers[z]] = got.Data[x][z]
			fmt.Println(entries[got.Headers[z]])
		}
	}
	// make sure we got the same number of headers back we expect.
	if len(got.Headers) != len(headers) {
		t.Fatalf("want %+v, got %+v", headers, got.Headers)
	}
	// create the data map
	for z := 0; z < len(got.Data[0]); z++ {
		entries[got.Headers[z]] = got.Data[0][z]
		//fmt.Println(entries[got.Headers[z]])
	}
	// verify the data map
	for key := range parsed {
		if parsed[key] != entries[key] {
			t.Fatalf("want %+v, got %+v", parsed[key], entries[key])
		}
	}
	// verify the tag map
	for key := range tags {
		if test_tags[key] != tags[key] {
			t.Fatalf("want %+v, got %+v", test_tags[key], tags[key])
		}
	}
}
