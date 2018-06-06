package syslog

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
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
					"timestamp":     time.Unix(1514764799, 999999000).UnixNano(),
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
					"timestamp":     time.Unix(1514764799, 999999000).UnixNano(),
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

func newUDPSyslogReceiver(address string, bestEffort bool) *Syslog {
	return &Syslog{
		Address: address,
		now: func() time.Time {
			return defaultTime
		},
		BestEffort: bestEffort,
		Separator:  "_",
	}
}

func testRFC5426(t *testing.T, protocol string, address string, bestEffort bool) {
	for _, tc := range getTestCasesForRFC5426() {
		t.Run(tc.name, func(t *testing.T) {
			// Create receiver
			receiver := newUDPSyslogReceiver(protocol+"://"+address, bestEffort)
			acc := &testutil.Accumulator{}
			require.NoError(t, receiver.Start(acc))
			defer receiver.Stop()

			// Clear
			acc.ClearMetrics()
			acc.Errors = make([]error, 0)

			// Connect
			conn, err := net.Dial(protocol, address)
			require.NotNil(t, conn)
			defer conn.Close()
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
	testRFC5426(t, "udp", address, true)
}

func TestStrict_udp(t *testing.T) {
	testRFC5426(t, "udp", address, false)
}

func TestBestEffort_unixgram(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestBestEffort_unixgram.sock")
	os.Create(sock)
	testRFC5426(t, "unixgram", sock, true)
}

func TestStrict_unixgram(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestStrict_unixgram.sock")
	os.Create(sock)
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
		Address:    "udp://" + address,
		now:        getNow,
		BestEffort: false,
		Separator:  "_",
	}
	acc := &testutil.Accumulator{}
	require.NoError(t, receiver.Start(acc))
	defer receiver.Stop()

	// Connect
	conn, err := net.Dial("udp", address)
	require.NotNil(t, conn)
	defer conn.Close()
	require.Nil(t, err)

	// Write
	_, e := conn.Write([]byte("<1>1 - - - - - -"))
	require.Nil(t, e)

	// Wait
	acc.Wait(1)

	want := &testutil.Metric{
		Measurement: "syslog",
		Fields: map[string]interface{}{
			"version":       uint16(1),
			"facility_code": 0,
			"severity_code": 1,
		},
		Tags: map[string]string{
			"severity": "alert",
			"facility": "kern",
		},
		Time: getNow(),
	}

	if !cmp.Equal(want, acc.Metrics[0]) {
		t.Fatalf("Got (+) / Want (-)\n %s", cmp.Diff(want, acc.Metrics[0]))
	}

	// New one with different time
	atomic.StoreInt64(&i, atomic.LoadInt64(&i)+1)

	// Clear
	acc.ClearMetrics()

	// Write
	_, e = conn.Write([]byte("<1>1 - - - - - -"))
	require.Nil(t, e)

	// Wait
	acc.Wait(1)

	want = &testutil.Metric{
		Measurement: "syslog",
		Fields: map[string]interface{}{
			"version":       uint16(1),
			"facility_code": 0,
			"severity_code": 1,
		},
		Tags: map[string]string{
			"severity": "alert",
			"facility": "kern",
		},
		Time: getNow(),
	}

	if !cmp.Equal(want, acc.Metrics[0]) {
		t.Fatalf("Got (+) / Want (-)\n %s", cmp.Diff(want, acc.Metrics[0]))
	}

	// New one with same time as previous one

	// Clear
	acc.ClearMetrics()

	// Write
	_, e = conn.Write([]byte("<1>1 - - - - - -"))
	require.Nil(t, e)

	// Wait
	acc.Wait(1)

	want = &testutil.Metric{
		Measurement: "syslog",
		Fields: map[string]interface{}{
			"version":       uint16(1),
			"facility_code": 0,
			"severity_code": 1,
		},
		Tags: map[string]string{
			"severity": "alert",
			"facility": "kern",
		},
		Time: getNow().Add(time.Nanosecond),
	}

	if !cmp.Equal(want, acc.Metrics[0]) {
		t.Fatalf("Got (+) / Want (-)\n %s", cmp.Diff(want, acc.Metrics[0]))
	}
}
