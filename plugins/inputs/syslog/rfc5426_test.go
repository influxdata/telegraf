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
		{
			name: "empty",
			data: []byte(""),
			werr: true,
		},
		{
			name: "complete",
			data: []byte("<1>1 - - - - - - A"),
			wantBestEffort: &testutil.Metric{
				Measurement: "syslog",
				Fields: map[string]interface{}{
					"version":       uint16(1),
					"message":       "A",
					"facility_code": 0,
					"severity_code": 1,
				},
				Tags: map[string]string{
					"severity": "alert",
					"facility": "kern",
				},
				Time: defaultTime,
			},
			wantStrict: &testutil.Metric{
				Measurement: "syslog",
				Fields: map[string]interface{}{
					"version":       uint16(1),
					"message":       "A",
					"facility_code": 0,
					"severity_code": 1,
				},
				Tags: map[string]string{
					"severity": "alert",
					"facility": "kern",
				},
				Time: defaultTime,
			},
		},
		{
			name: "one/per/packet",
			data: []byte("<1>3 - - - - - - A<1>4 - - - - - - B"),
			wantBestEffort: &testutil.Metric{
				Measurement: "syslog",
				Fields: map[string]interface{}{
					"version":       uint16(3),
					"message":       "A<1>4 - - - - - - B",
					"severity_code": 1,
					"facility_code": 0,
				},
				Tags: map[string]string{
					"severity": "alert",
					"facility": "kern",
				},
				Time: defaultTime,
			},
			wantStrict: &testutil.Metric{
				Measurement: "syslog",
				Fields: map[string]interface{}{
					"version":       uint16(3),
					"message":       "A<1>4 - - - - - - B",
					"severity_code": 1,
					"facility_code": 0,
				},
				Tags: map[string]string{
					"severity": "alert",
					"facility": "kern",
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
					"severity_code": 5,
					"facility_code": 3,
				},
				Tags: map[string]string{
					"severity": "notice",
					"facility": "daemon",
					"hostname": "web1",
					"appname":  "someservice",
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
					"severity_code": 5,
					"facility_code": 3,
				},
				Tags: map[string]string{
					"severity": "notice",
					"facility": "daemon",
					"hostname": "web1",
					"appname":  "someservice",
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
					"version":       maxV,
					"timestamp":     time.Unix(1514764799, 999999000).UTC(),
					"message":       message7681,
					"procid":        maxPID,
					"msgid":         maxMID,
					"severity_code": 7,
					"facility_code": 23,
				},
				Tags: map[string]string{
					"severity": "debug",
					"facility": "local7",
					"hostname": maxH,
					"appname":  maxA,
				},
				Time: defaultTime,
			},
			wantStrict: &testutil.Metric{
				Measurement: "syslog",
				Fields: map[string]interface{}{
					"version":       maxV,
					"timestamp":     time.Unix(1514764799, 999999000).UTC(),
					"message":       message7681,
					"procid":        maxPID,
					"msgid":         maxMID,
					"severity_code": 7,
					"facility_code": 23,
				},
				Tags: map[string]string{
					"severity": "debug",
					"facility": "local7",
					"hostname": maxH,
					"appname":  maxA,
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
					"version":       uint16(2),
					"facility_code": 0,
					"severity_code": 1,
				},
				Tags: map[string]string{
					"severity": "alert",
					"facility": "kern",
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

func testRFC5426(t *testing.T, bestEffort bool) {
	for _, tc := range getTestCasesForRFC5426() {
		t.Run(tc.name, func(t *testing.T) {
			// Create receiver
			receiver := newUDPSyslogReceiver(bestEffort)
			require.Equal(t, receiver.Protocol, "udp")
			acc := &testutil.Accumulator{}
			require.NoError(t, receiver.Start(acc))
			defer receiver.Stop()

			// Clear
			acc.ClearMetrics()
			acc.Errors = make([]error, 0)

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

func TestBestEffort_udp(t *testing.T) {
	testRFC5426(t, true)
}

func TestStrict_udp(t *testing.T) {
	testRFC5426(t, false)
}
