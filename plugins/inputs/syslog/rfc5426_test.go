package syslog

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

type testCase5426 struct {
	name           string
	data           []byte
	wantBestEffort *testutil.Metric
	wantStrict     *testutil.Metric
	werr           bool
}

func getTestCasesForRFC5426() []testCase5426 {
	testCases := []testCase5426{
		// (fixme) > need a timeout
		// {
		// 	name: "empty",
		// 	data: []byte(""),
		// },
		{
			name: "complete",
			data: []byte("<1>1 - - - - - - A"),
			wantBestEffort: &testutil.Metric{
				Measurement: "syslog",
				Fields: map[string]interface{}{
					"version": uint16(1),
					"message": "A",
				},
				Tags: map[string]string{
					"severity":         "1",
					"severity_level":   "alert",
					"facility":         "0",
					"facility_message": "kernel messages",
				},
				Time: defaultTime,
			},
			wantStrict: &testutil.Metric{
				Measurement: "syslog",
				Fields: map[string]interface{}{
					"version": uint16(1),
					"message": "A",
				},
				Tags: map[string]string{
					"severity":         "1",
					"severity_level":   "alert",
					"facility":         "0",
					"facility_message": "kernel messages",
				},
				Time: defaultTime,
			},
		},
		{
			name: "average",
			data: []byte(`<29>1 2016-02-21T04:32:57+00:00 web1 someservice 2341 2 [origin][meta sequence="14125553" service="someservice"] "GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`),
			wantBestEffort: &testutil.Metric{
				Measurement: "syslog",
				Fields: map[string]interface{}{
					"version":       uint16(1),
					"timestamp":     time.Unix(1456029177, 0).UTC(),
					"procid":        "2341",
					"msgid":         "2",
					"message":       `"GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`,
					"origin":        true,
					"meta sequence": "14125553",
					"meta service":  "someservice",
				},
				Tags: map[string]string{
					"severity":         "5",
					"severity_level":   "notice",
					"facility":         "3",
					"facility_message": "system daemons",
					"hostname":         "web1",
					"appname":          "someservice",
				},
				Time: defaultTime,
			},
			wantStrict: &testutil.Metric{
				Measurement: "syslog",
				Fields: map[string]interface{}{
					"version":       uint16(1),
					"timestamp":     time.Unix(1456029177, 0).UTC(),
					"procid":        "2341",
					"msgid":         "2",
					"message":       `"GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`,
					"origin":        true,
					"meta sequence": "14125553",
					"meta service":  "someservice",
				},
				Tags: map[string]string{
					"severity":         "5",
					"severity_level":   "notice",
					"facility":         "3",
					"facility_message": "system daemons",
					"hostname":         "web1",
					"appname":          "someservice",
				},
				Time: defaultTime,
			},
		},
		{
			name: "max",
			data: []byte(fmt.Sprintf("<%d>%d %s %s %s %s %s - %s", maxP, maxV, maxTS, maxH, maxA, maxPID, maxMID, message7681)),
			wantBestEffort: &testutil.Metric{
				Measurement: "syslog",
				Fields: map[string]interface{}{
					"version":   maxV,
					"timestamp": time.Unix(1514764799, 999999000).UTC(),
					"message":   message7681,
					"procid":    maxPID,
					"msgid":     maxMID,
				},
				Tags: map[string]string{
					"severity":         "7",
					"severity_level":   "debug",
					"facility":         "23",
					"facility_message": "local use 7 (local7)",
					"hostname":         maxH,
					"appname":          maxA,
				},
				Time: defaultTime,
			},
			wantStrict: &testutil.Metric{
				Measurement: "syslog",
				Fields: map[string]interface{}{
					"version":   maxV,
					"timestamp": time.Unix(1514764799, 999999000).UTC(),
					"message":   message7681,
					"procid":    maxPID,
					"msgid":     maxMID,
				},
				Tags: map[string]string{
					"severity":         "7",
					"severity_level":   "debug",
					"facility":         "23",
					"facility_message": "local use 7 (local7)",
					"hostname":         maxH,
					"appname":          maxA,
				},
				Time: defaultTime,
			},
		},
		{
			name: "minimal/incomplete",
			data: []byte("<1>2"),
			wantBestEffort: &testutil.Metric{
				Measurement: "syslog",
				Fields: map[string]interface{}{
					"version": uint16(2),
				},
				Tags: map[string]string{
					"severity":         "1",
					"severity_level":   "alert",
					"facility":         "0",
					"facility_message": "kernel messages",
				},
				Time: defaultTime,
			},
			werr: true,
		},
	}

	return testCases
}

func newUDPSyslogReceiver(bestEffort bool) *Syslog {
	return &Syslog{
		Protocol: "udp",
		Address:  address,
		now: func() time.Time {
			return defaultTime
		},
		BestEffort: bestEffort,
	}
}

func testRFC5426(t *testing.T, acc *testutil.Accumulator, bestEffort bool) {
	for _, tc := range getTestCasesForRFC5426() {
		t.Run(tc.name, func(t *testing.T) {
			// Clear
			acc.ClearMetrics()
			// Connect
			conn, err := net.Dial("udp", address)
			defer conn.Close()
			require.NotNil(t, conn)
			require.Nil(t, err)
			// Write
			_, e := conn.Write(tc.data)
			require.Nil(t, e)
			// Waiting ...
			if tc.wantStrict == nil && tc.werr || bestEffort && tc.werr {
				acc.WaitError(1)
			}
			if tc.wantBestEffort != nil && bestEffort || tc.wantStrict != nil && !bestEffort {
				acc.Wait(1) // RFC5426 mandates a syslog message per UDP packet
			}
			// Compare
			var got *testutil.Metric
			var want *testutil.Metric
			if len(acc.Metrics) > 0 {
				got = acc.Metrics[0]
			}
			if bestEffort {
				want = tc.wantBestEffort
			} else {
				want = tc.wantStrict
			}
			if !cmp.Equal(want, got) {
				t.Fatalf("Got (+) / Want (-)\n %s", cmp.Diff(want, got))
			}
		})
	}
}

func TestUDPInBestEffortMode(t *testing.T) {
	bestEffort := true
	receiver := newUDPSyslogReceiver(bestEffort)
	require.Equal(t, receiver.Protocol, "udp")

	acc := &testutil.Accumulator{}
	require.NoError(t, receiver.Start(acc))
	defer receiver.Stop()

	testRFC5426(t, acc, bestEffort)
}

func TestUDPInStrictMode(t *testing.T) {
	receiver := newUDPSyslogReceiver(false)
	require.Equal(t, receiver.Protocol, "udp")

	acc := &testutil.Accumulator{}
	require.NoError(t, receiver.Start(acc))
	defer receiver.Stop()

	testRFC5426(t, acc, false)
}
