package chronyc

import (
	// "fmt"
	//	"os"
	"github.com/influxdata/telegraf/testutil"
	"reflect"
	"strings"
	"testing"
)

func TestParseTracking(t *testing.T) {

	line := "50505300,PPS,1,1541793798.895264285,-0.000001007,0.000000291,0.000000239,-17.957,0.000,0.005,0.000001000,0.000010123,16.0,Normal"
	rTags := map[string]string{
		"command": "tracking",
		"clockId": "chrony",
	}
	rFields := map[string]interface{}{
		"refId":            "PPS",
		"refIdHex":         "50505300",
		"stratum":          int64(1),
		"refTime":          1541793798.895264285,
		"systemTimeOffset": -0.000001007,
		"lastOffset":       0.000000291,
		"rmsOffset":        0.000000239,
		"frequency":        -17.957,
		"freqResidual":     0.000,
		"freqSkew":         0.005,
		"rootDelay":        0.000001000,
		"rootDispersion":   0.000010123,
		"updateInterval":   16.0,
		"leapStatus":       "Normal",
	}

	tFields, tTags, err := parseTracking(strings.Split(line, ","))

	if err != nil {
		t.Fatalf("An error has been returned: %v", err)
	}
	if !reflect.DeepEqual(tFields, rFields) {
		t.Fatalf("Fields not equal:\n want %#v,\n have %#v", rFields, tFields)
	}
	if !reflect.DeepEqual(tTags, rTags) {
		t.Fatalf("Tags not equal:\n want %#v,\n have %#v", rTags, tTags)
	}
}

func TestParseSources(t *testing.T) {

	commandList := []string{"sources"}
	out :=
		`#,?,NMEA,0,4,377,11,-0.000734182,-0.000734058,0.100231066
#,*,PPS,0,4,377,12,0.000000701,0.000000825,0.000000684
^,-,194.190.168.1,15,10,377,20,-0.000373457,-0.000373289,0.001973625
^,?,10.10.10.10,0,10,0,4294967295,0.000000000,0.000000000,0.000000000
`
	var acc testutil.Accumulator
	c := Chrony{}

	err := c.parseChronycOutput(commandList, out, &acc)

	if err != nil {
		t.Fatalf("An error has been returned: %v", err)
	}

	if acc.NMetrics() != 4 {
		t.Fatalf("%d metrics total, must be 4", acc.NMetrics())
	}

	rFields := map[string]interface{}{
		"clockMode":   2,
		"clockState":  1,
		"stratum":     int64(0),
		"poll":        int64(4),
		"reach":       int64(255),
		"lastRx":      int64(11),
		"offset":      -0.000734182,
		"rawOffset":   -0.000734058,
		"errorMargin": 0.100231066,
	}
	rTags := map[string]string{
		"command": "sources",
		"clockId": "NMEA",
	}
	acc.AssertContainsTaggedFields(t, "chronyc", rFields, rTags)

	rFields = map[string]interface{}{
		"clockMode":   2,
		"clockState":  0,
		"stratum":     int64(0),
		"poll":        int64(4),
		"reach":       int64(255),
		"lastRx":      int64(12),
		"offset":      0.000000701,
		"rawOffset":   0.000000825,
		"errorMargin": 0.000000684,
	}
	rTags = map[string]string{
		"command": "sources",
		"clockId": "PPS",
	}
	acc.AssertContainsTaggedFields(t, "chronyc", rFields, rTags)

	rFields = map[string]interface{}{
		"stratum":     int64(15),
		"reach":       int64(255),
		"lastRx":      int64(20),
		"clockMode":   0,
		"offset":      -0.000373457,
		"errorMargin": 0.001973625,
		"rawOffset":   -0.000373289,
		"clockState":  5,
		"poll":        int64(10),
	}
	rTags = map[string]string{
		"command": "sources",
		"clockId": "194.190.168.1",
	}
	acc.AssertContainsTaggedFields(t, "chronyc", rFields, rTags)

	rFields = map[string]interface{}{
		"poll":       int64(10),
		"reach":      int64(0),
		"clockMode":  0,
		"clockState": 1,
	}
	rTags = map[string]string{
		"command": "sources",
		"clockId": "10.10.10.10",
	}
	acc.AssertContainsTaggedFields(t, "chronyc", rFields, rTags)

}

func TestParseNoSources(t *testing.T) {

	commandList := []string{"tracking", "sources", "serverstats"}
	out :=
		`50505300,PPS,1,1542141310.995435528,0.000000132,0.000000013,0.000000059,-17.935,0.000,0.001,0.000001000,0.000019609,16.0,Normal
13,0,464070,0,0
`

	var acc testutil.Accumulator
	c := Chrony{}

	err := c.parseChronycOutput(commandList, out, &acc)

	if err != nil {
		t.Fatalf("An error has been returned: %v", err)
	}

	if acc.NMetrics() != 2 {
		t.Fatalf("%d metrics total, must be 2", acc.NMetrics())
	}
}

func TestDoubleTracking(t *testing.T) {

	commandList := []string{"tracking", "tracking"}
	out :=
		`50505300,PPS,1,1542141310.995435528,0.000000132,0.000000013,0.000000059,-17.935,0.000,0.001,0.000001000,0.000019609,16.0,Normal
50505300,PPS,1,1542141310.995435528,0.000000132,0.000000013,0.000000059,-17.935,0.000,0.001,0.000001000,0.000019609,16.0,Normal
`

	var acc testutil.Accumulator
	c := Chrony{}

	err := c.parseChronycOutput(commandList, out, &acc)

	if err != nil {
		t.Fatalf("An error has been returned: %v", err)
	}

	t.Logf("acc: %v", &acc)

	if acc.NMetrics() != 2 {
		t.Fatalf("%d metrics total, must be 2", acc.NMetrics())
	}
}

func TestParseClients(t *testing.T) {

	commandList := []string{"clients"}
	out :=
		`127.0.0.1,123,22,127,127,4294967295,83748,0,3,14
185.245.86.226,1,0,127,127,819881,0,0,127,4294967295
54.211.253.78,1,0,127,127,817691,0,0,127,4294967295
`
	var acc testutil.Accumulator
	c := Chrony{}

	err := c.parseChronycOutput(commandList, out, &acc)

	if err != nil {
		t.Fatalf("An error has been returned: %v", err)
	}

	if acc.NMetrics() != 4 {
		t.Fatalf("%d metrics total, must be 4 (including summary)", acc.NMetrics())
	}

	t.Logf("acc: %v", &acc)

	rFields := map[string]interface{}{
		"ntpRequests":    int64(123),
		"ntpDropped":     int64(22),
		"cmdInterval":    int64(3),
		"cmdLastRequest": int64(14),
		"cmdRequests":    int64(83748),
		"cmdDropped":     int64(0),
	}
	rTags := map[string]string{
		"command":       "clients",
		"clientAddress": "127.0.0.1",
	}
	acc.AssertContainsTaggedFields(t, "chronyc", rFields, rTags)

}

func TestParseClientsSummary(t *testing.T) {

	commandList := []string{"clients"}
	out :=
		`127.0.0.1,0,0,127,127,4294967295,83748,0,3,0
185.245.86.226,1,0,127,127,819881,0,0,127,4294967295
54.211.253.78,1,0,127,127,817691,0,0,127,4294967295
`
	var acc testutil.Accumulator
	c := Chrony{
		ClientsSummary: true,
	}

	err := c.parseChronycOutput(commandList, out, &acc)

	if err != nil {
		t.Fatalf("An error has been returned: %v", err)
	}

	t.Logf("acc: %v", &acc)

	if acc.NMetrics() != 1 {
		t.Fatalf("%d metrics total, must be 1 (summary only)", acc.NMetrics())
	}

	rFields := map[string]interface{}{
		"ntpClients":       int64(2),
		"activeNtpClients": int64(0),
		"totalClients":     int64(3),
	}
	rTags := map[string]string{
		"command": "clients",
	}
	acc.AssertContainsTaggedFields(t, "chronyc", rFields, rTags)

}
