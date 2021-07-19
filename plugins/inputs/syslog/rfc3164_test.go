package syslog

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func timeMustParse(value string) time.Time {
	format := "Jan 2 15:04:05 2006"
	t, err := time.Parse(format, value)
	if err != nil {
		panic(fmt.Sprintf("couldn't parse time: %v", value))
	}
	return t
}

func getTestCasesForRFC3164() []testCasePacket {
	currentYear := time.Now().Year()
	ts := timeMustParse(fmt.Sprintf("Dec 2 16:31:03 %d", currentYear)).UnixNano()
	testCases := []testCasePacket{
		{
			name: "complete",
			data: []byte("<13>Dec  2 16:31:03 host app: Test"),
			wantBestEffort: testutil.MustMetric(
				"syslog",
				map[string]string{
					"appname":  "app",
					"severity": "notice",
					"hostname": "host",
					"facility": "user",
				},
				map[string]interface{}{
					"timestamp":     ts,
					"message":       "Test",
					"facility_code": 1,
					"severity_code": 5,
				},
				defaultTime,
			),
			wantStrict: testutil.MustMetric(
				"syslog",
				map[string]string{
					"appname":  "app",
					"severity": "notice",
					"hostname": "host",
					"facility": "user",
				},
				map[string]interface{}{
					"timestamp":     ts,
					"message":       "Test",
					"facility_code": 1,
					"severity_code": 5,
				},
				defaultTime,
			),
		},
	}

	return testCases
}

func testRFC3164(t *testing.T, protocol string, address string, bestEffort bool) {
	for _, tc := range getTestCasesForRFC3164() {
		t.Run(tc.name, func(t *testing.T) {
			// Create receiver
			receiver := newUDPSyslogReceiver(protocol+"://"+address, bestEffort, syslogRFC3164)
			acc := &testutil.Accumulator{}
			require.NoError(t, receiver.Start(acc))
			defer receiver.Stop()

			// Connect
			conn, err := net.Dial(protocol, address)
			require.NotNil(t, conn)
			require.NoError(t, err)

			// Write
			_, err = conn.Write(tc.data)
			conn.Close()
			if err != nil {
				if err, ok := err.(*net.OpError); ok {
					if err.Err.Error() == "write: message too long" {
						return
					}
				}
			}

			// Waiting ...
			if tc.wantStrict == nil && tc.werr || bestEffort && tc.werr {
				acc.WaitError(1)
			}
			if tc.wantBestEffort != nil && bestEffort || tc.wantStrict != nil && !bestEffort {
				acc.Wait(1) // RFC3164 mandates a syslog message per UDP packet
			}

			// Compare
			var got telegraf.Metric
			var want telegraf.Metric
			if len(acc.Metrics) > 0 {
				got = acc.GetTelegrafMetrics()[0]
			}
			if bestEffort {
				want = tc.wantBestEffort
			} else {
				want = tc.wantStrict
			}
			testutil.RequireMetricEqual(t, want, got)
		})
	}
}

func TestRFC3164BestEffort_udp(t *testing.T) {
	testRFC3164(t, "udp", address, true)
}

func TestRFC3164Strict_udp(t *testing.T) {
	testRFC3164(t, "udp", address, false)
}
