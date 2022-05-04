package syslog

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

func getTestCasesForRFC5426() []testCasePacket {
	testCases := []testCasePacket{
		{
			name: "complete",
			data: []byte("<1>1 - - - - - - A"),
			wantBestEffort: testutil.MustMetric(
				"syslog",
				map[string]string{
					"severity": "alert",
					"facility": "kern",
				},
				map[string]interface{}{
					"version":       uint16(1),
					"message":       "A",
					"facility_code": 0,
					"severity_code": 1,
				},
				defaultTime,
			),
			wantStrict: testutil.MustMetric(
				"syslog",
				map[string]string{
					"severity": "alert",
					"facility": "kern",
				},
				map[string]interface{}{
					"version":       uint16(1),
					"message":       "A",
					"facility_code": 0,
					"severity_code": 1,
				},
				defaultTime,
			),
		},
		{
			name: "one/per/packet",
			data: []byte("<1>3 - - - - - - A<1>4 - - - - - - B"),
			wantBestEffort: testutil.MustMetric(
				"syslog",
				map[string]string{
					"severity": "alert",
					"facility": "kern",
				},
				map[string]interface{}{
					"version":       uint16(3),
					"message":       "A<1>4 - - - - - - B",
					"severity_code": 1,
					"facility_code": 0,
				},
				defaultTime,
			),
			wantStrict: testutil.MustMetric(
				"syslog",
				map[string]string{
					"severity": "alert",
					"facility": "kern",
				},
				map[string]interface{}{
					"version":       uint16(3),
					"message":       "A<1>4 - - - - - - B",
					"severity_code": 1,
					"facility_code": 0,
				},
				defaultTime,
			),
		},
		{
			name: "average",
			data: []byte(`<29>1 2016-02-21T04:32:57+00:00 web1 someservice 2341 2 [origin][meta sequence="14125553" service="someservice"] "GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`),
			wantBestEffort: testutil.MustMetric(
				"syslog",
				map[string]string{
					"severity": "notice",
					"facility": "daemon",
					"hostname": "web1",
					"appname":  "someservice",
				},
				map[string]interface{}{
					"version":       uint16(1),
					"timestamp":     time.Unix(1456029177, 0).UnixNano(),
					"procid":        "2341",
					"msgid":         "2",
					"message":       `"GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`,
					"origin":        true,
					"meta_sequence": "14125553",
					"meta_service":  "someservice",
					"severity_code": 5,
					"facility_code": 3,
				},
				defaultTime,
			),
			wantStrict: testutil.MustMetric(
				"syslog",
				map[string]string{
					"severity": "notice",
					"facility": "daemon",
					"hostname": "web1",
					"appname":  "someservice",
				},
				map[string]interface{}{
					"version":       uint16(1),
					"timestamp":     time.Unix(1456029177, 0).UnixNano(),
					"procid":        "2341",
					"msgid":         "2",
					"message":       `"GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`,
					"origin":        true,
					"meta_sequence": "14125553",
					"meta_service":  "someservice",
					"severity_code": 5,
					"facility_code": 3,
				},
				defaultTime,
			),
		},
		{
			name: "max",
			data: []byte(fmt.Sprintf("<%d>%d %s %s %s %s %s - %s", maxP, maxV, maxTS, maxH, maxA, maxPID, maxMID, message7681)),
			wantBestEffort: testutil.MustMetric(
				"syslog",
				map[string]string{
					"severity": "debug",
					"facility": "local7",
					"hostname": maxH,
					"appname":  maxA,
				},
				map[string]interface{}{
					"version":       maxV,
					"timestamp":     time.Unix(1514764799, 999999000).UnixNano(),
					"message":       message7681,
					"procid":        maxPID,
					"msgid":         maxMID,
					"severity_code": 7,
					"facility_code": 23,
				},
				defaultTime,
			),
			wantStrict: testutil.MustMetric(
				"syslog",
				map[string]string{
					"severity": "debug",
					"facility": "local7",
					"hostname": maxH,
					"appname":  maxA,
				},
				map[string]interface{}{
					"version":       maxV,
					"timestamp":     time.Unix(1514764799, 999999000).UnixNano(),
					"message":       message7681,
					"procid":        maxPID,
					"msgid":         maxMID,
					"severity_code": 7,
					"facility_code": 23,
				},
				defaultTime,
			),
		},
		{
			name: "minimal/incomplete",
			data: []byte("<1>2"),
			wantBestEffort: testutil.MustMetric(
				"syslog",
				map[string]string{
					"severity": "alert",
					"facility": "kern",
				},
				map[string]interface{}{
					"version":       uint16(2),
					"facility_code": 0,
					"severity_code": 1,
				},
				defaultTime,
			),
			werr: true,
		},
		{
			name: "trim message",
			data: []byte("<1>1 - - - - - - \tA\n"),
			wantBestEffort: testutil.MustMetric(
				"syslog",
				map[string]string{
					"severity": "alert",
					"facility": "kern",
				},
				map[string]interface{}{
					"version":       uint16(1),
					"message":       "\tA",
					"facility_code": 0,
					"severity_code": 1,
				},
				defaultTime,
			),
			wantStrict: testutil.MustMetric(
				"syslog",
				map[string]string{
					"severity": "alert",
					"facility": "kern",
				},
				map[string]interface{}{
					"version":       uint16(1),
					"message":       "\tA",
					"facility_code": 0,
					"severity_code": 1,
				},
				defaultTime,
			),
		},
	}

	return testCases
}

func testRFC5426(t *testing.T, protocol string, address string, bestEffort bool) {
	for _, tc := range getTestCasesForRFC5426() {
		t.Run(tc.name, func(t *testing.T) {
			// Create receiver
			receiver := newUDPSyslogReceiver(protocol+"://"+address, bestEffort, syslogRFC5424)
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
				acc.Wait(1) // RFC5426 mandates a syslog message per UDP packet
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

func TestBestEffort_udp(t *testing.T) {
	testRFC5426(t, "udp", address, true)
}

func TestStrict_udp(t *testing.T) {
	testRFC5426(t, "udp", address, false)
}

func TestBestEffort_unixgram(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows, as unixgram sockets are not supported")
	}

	sock := testutil.TempSocket(t)
	f, err := os.Create(sock)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, f.Close()) })

	testRFC5426(t, "unixgram", sock, true)
}

func TestStrict_unixgram(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows, as unixgram sockets are not supported")
	}

	sock := testutil.TempSocket(t)
	f, err := os.Create(sock)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, f.Close()) })

	testRFC5426(t, "unixgram", sock, false)
}

func TestTimeIncrement_udp(t *testing.T) {
	var i int64
	atomic.StoreInt64(&i, 0)
	getNow := func() time.Time {
		if atomic.LoadInt64(&i)%2 == 0 {
			return time.Unix(1, 0)
		}
		return time.Unix(1, 1)
	}

	// Create receiver
	receiver := &Syslog{
		Address:        "udp://" + address,
		now:            getNow,
		BestEffort:     false,
		SyslogStandard: syslogRFC5424,
		Separator:      "_",
	}
	acc := &testutil.Accumulator{}
	require.NoError(t, receiver.Start(acc))
	defer receiver.Stop()

	// Connect
	conn, err := net.Dial("udp", address)
	require.NotNil(t, conn)
	defer conn.Close()
	require.NoError(t, err)

	// Write
	_, e := conn.Write([]byte("<1>1 - - - - - -"))
	require.Nil(t, e)

	// Wait
	acc.Wait(1)

	want := []telegraf.Metric{
		testutil.MustMetric(
			"syslog",
			map[string]string{
				"severity": "alert",
				"facility": "kern",
			},
			map[string]interface{}{
				"version":       uint16(1),
				"facility_code": 0,
				"severity_code": 1,
			},
			getNow(),
		),
	}
	testutil.RequireMetricsEqual(t, want, acc.GetTelegrafMetrics())

	// New one with different time
	atomic.StoreInt64(&i, atomic.LoadInt64(&i)+1)

	// Clear
	acc.ClearMetrics()

	// Write
	_, e = conn.Write([]byte("<1>1 - - - - - -"))
	require.Nil(t, e)

	// Wait
	acc.Wait(1)

	want = []telegraf.Metric{
		testutil.MustMetric(
			"syslog",
			map[string]string{
				"severity": "alert",
				"facility": "kern",
			},
			map[string]interface{}{
				"version":       uint16(1),
				"facility_code": 0,
				"severity_code": 1,
			},
			getNow(),
		),
	}
	testutil.RequireMetricsEqual(t, want, acc.GetTelegrafMetrics())

	// New one with same time as previous one

	// Clear
	acc.ClearMetrics()

	// Write
	_, e = conn.Write([]byte("<1>1 - - - - - -"))
	require.Nil(t, e)

	// Wait
	acc.Wait(1)

	want = []telegraf.Metric{
		testutil.MustMetric(
			"syslog",
			map[string]string{
				"severity": "alert",
				"facility": "kern",
			},
			map[string]interface{}{
				"version":       uint16(1),
				"facility_code": 0,
				"severity_code": 1,
			},
			getNow().Add(time.Nanosecond),
		),
	}
	testutil.RequireMetricsEqual(t, want, acc.GetTelegrafMetrics())
}
