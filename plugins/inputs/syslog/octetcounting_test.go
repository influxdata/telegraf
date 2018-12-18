package syslog

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

func getTestCasesForOctetCounting() []testCaseStream {
	testCases := []testCaseStream{
		{
			name: "1st/avg/ok",
			data: []byte(`188 <29>1 2016-02-21T04:32:57+00:00 web1 someservice 2341 2 [origin][meta sequence="14125553" service="someservice"] "GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`),
			wantStrict: []testutil.Metric{
				{
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
			wantBestEffort: []testutil.Metric{
				{
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
		},
		{
			name: "1st/min/ok//2nd/min/ok",
			data: []byte("16 <1>2 - - - - - -17 <4>11 - - - - - -"),
			wantStrict: []testutil.Metric{
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       uint16(2),
						"severity_code": 1,
						"facility_code": 0,
					},
					Tags: map[string]string{
						"severity": "alert",
						"facility": "kern",
					},
					Time: defaultTime,
				},
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       uint16(11),
						"severity_code": 4,
						"facility_code": 0,
					},
					Tags: map[string]string{
						"severity": "warning",
						"facility": "kern",
					},
					Time: defaultTime.Add(time.Nanosecond),
				},
			},
			wantBestEffort: []testutil.Metric{
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       uint16(2),
						"severity_code": 1,
						"facility_code": 0,
					},
					Tags: map[string]string{
						"severity": "alert",
						"facility": "kern",
					},
					Time: defaultTime,
				},
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       uint16(11),
						"severity_code": 4,
						"facility_code": 0,
					},
					Tags: map[string]string{
						"severity": "warning",
						"facility": "kern",
					},
					Time: defaultTime.Add(time.Nanosecond),
				},
			},
		},
		{
			name: "1st/utf8/ok",
			data: []byte("23 <1>1 - - - - - - hellø"),
			wantStrict: []testutil.Metric{
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       uint16(1),
						"message":       "hellø",
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
			wantBestEffort: []testutil.Metric{
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       uint16(1),
						"message":       "hellø",
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
		},
		{
			name: "1st/nl/ok", // newline
			data: []byte("28 <1>3 - - - - - - hello\nworld"),
			wantStrict: []testutil.Metric{
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       uint16(3),
						"message":       "hello\nworld",
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
			wantBestEffort: []testutil.Metric{
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       uint16(3),
						"message":       "hello\nworld",
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
		},
		{
			name:       "1st/uf/ko", // underflow (msglen less than provided octets)
			data:       []byte("16 <1>2"),
			wantStrict: nil,
			wantBestEffort: []testutil.Metric{
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       uint16(2),
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
			werr: 1,
		},
		{
			name: "1st/min/ok",
			data: []byte("16 <1>1 - - - - - -"),
			wantStrict: []testutil.Metric{
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       uint16(1),
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
			wantBestEffort: []testutil.Metric{
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       uint16(1),
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
		},
		{
			name:       "1st/uf/mf", // The first "underflow" message breaks also the second one
			data:       []byte("16 <1>217 <11>1 - - - - - -"),
			wantStrict: nil,
			wantBestEffort: []testutil.Metric{
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       uint16(217),
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
			werr: 1,
		},
		// {
		// 	name: "1st/of/ko", // overflow (msglen greather then max allowed octets)
		// 	data: []byte(fmt.Sprintf("8193 <%d>%d %s %s %s %s %s 12 %s", maxP, maxV, maxTS, maxH, maxA, maxPID, maxMID, message7681)),
		// 	want: []testutil.Metric{},
		// },
		{
			name: "1st/max/ok",
			data: []byte(fmt.Sprintf("8192 <%d>%d %s %s %s %s %s - %s", maxP, maxV, maxTS, maxH, maxA, maxPID, maxMID, message7681)),
			wantStrict: []testutil.Metric{
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       maxV,
						"timestamp":     time.Unix(1514764799, 999999000).UnixNano(),
						"message":       message7681,
						"procid":        maxPID,
						"msgid":         maxMID,
						"facility_code": 23,
						"severity_code": 7,
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
			wantBestEffort: []testutil.Metric{
				{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version":       maxV,
						"timestamp":     time.Unix(1514764799, 999999000).UnixNano(),
						"message":       message7681,
						"procid":        maxPID,
						"msgid":         maxMID,
						"facility_code": 23,
						"severity_code": 7,
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
		},
	}

	return testCases
}

func testStrictOctetCounting(t *testing.T, protocol string, address string, wantTLS bool, keepAlive *internal.Duration) {
	for _, tc := range getTestCasesForOctetCounting() {
		t.Run(tc.name, func(t *testing.T) {
			// Creation of a strict mode receiver
			receiver := newTCPSyslogReceiver(protocol+"://"+address, keepAlive, 0, false, OctetCounting)
			require.NotNil(t, receiver)
			if wantTLS {
				receiver.ServerConfig = *pki.TLSServerConfig()
			}
			require.Equal(t, receiver.KeepAlivePeriod, keepAlive)
			acc := &testutil.Accumulator{}
			require.NoError(t, receiver.Start(acc))
			defer receiver.Stop()

			// Connect
			var conn net.Conn
			var err error
			if wantTLS {
				config, e := pki.TLSClientConfig().TLSConfig()
				require.NoError(t, e)
				config.ServerName = "localhost"
				conn, err = tls.Dial(protocol, address, config)
			} else {
				conn, err = net.Dial(protocol, address)
				defer conn.Close()
			}
			require.NotNil(t, conn)
			require.NoError(t, err)

			// Clear
			acc.ClearMetrics()
			acc.Errors = make([]error, 0)

			// Write
			_, err = conn.Write(tc.data)
			conn.Close()
			require.NoError(t, err)

			// Wait that the the number of data points is accumulated
			// Since the receiver is running concurrently
			if tc.wantStrict != nil {
				acc.Wait(len(tc.wantStrict))
			}

			// Wait the parsing error
			acc.WaitError(tc.werr)

			// Verify
			if len(acc.Errors) != tc.werr {
				t.Fatalf("Got unexpected errors. want error = %v, errors = %v\n", tc.werr, acc.Errors)
			}
			var got []testutil.Metric
			for _, metric := range acc.Metrics {
				got = append(got, *metric)
			}
			if !cmp.Equal(tc.wantStrict, got) {
				t.Fatalf("Got (+) / Want (-)\n %s", cmp.Diff(tc.wantStrict, got))
			}
		})
	}
}

func testBestEffortOctetCounting(t *testing.T, protocol string, address string, wantTLS bool, keepAlive *internal.Duration) {
	for _, tc := range getTestCasesForOctetCounting() {
		t.Run(tc.name, func(t *testing.T) {
			// Creation of a best effort mode receiver
			receiver := newTCPSyslogReceiver(protocol+"://"+address, keepAlive, 0, true, OctetCounting)
			require.NotNil(t, receiver)
			if wantTLS {
				receiver.ServerConfig = *pki.TLSServerConfig()
			}
			require.Equal(t, receiver.KeepAlivePeriod, keepAlive)
			acc := &testutil.Accumulator{}
			require.NoError(t, receiver.Start(acc))
			defer receiver.Stop()

			// Connect
			var conn net.Conn
			var err error
			if wantTLS {
				config, e := pki.TLSClientConfig().TLSConfig()
				require.NoError(t, e)
				config.ServerName = "localhost"
				conn, err = tls.Dial(protocol, address, config)
			} else {
				conn, err = net.Dial(protocol, address)
			}
			require.NotNil(t, conn)
			require.NoError(t, err)

			// Clear
			acc.ClearMetrics()
			acc.Errors = make([]error, 0)

			// Write
			_, err = conn.Write(tc.data)
			require.NoError(t, err)
			conn.Close()

			// Wait that the the number of data points is accumulated
			// Since the receiver is running concurrently
			if tc.wantBestEffort != nil {
				acc.Wait(len(tc.wantBestEffort))
			}

			// Verify
			var got []testutil.Metric
			for _, metric := range acc.Metrics {
				got = append(got, *metric)
			}
			if !cmp.Equal(tc.wantBestEffort, got) {
				t.Fatalf("Got (+) / Want (-)\n %s", cmp.Diff(tc.wantBestEffort, got))
			}
		})
	}
}

func TestOctetCountingStrict_tcp(t *testing.T) {
	testStrictOctetCounting(t, "tcp", address, false, nil)
}

func TestOctetCountingBestEffort_tcp(t *testing.T) {
	testBestEffortOctetCounting(t, "tcp", address, false, nil)
}

func TestOctetCountingStrict_tcp_tls(t *testing.T) {
	testStrictOctetCounting(t, "tcp", address, true, nil)
}

func TestOctetCountingBestEffort_tcp_tls(t *testing.T) {
	testBestEffortOctetCounting(t, "tcp", address, true, nil)
}

func TestOctetCountingStrictWithKeepAlive_tcp_tls(t *testing.T) {
	testStrictOctetCounting(t, "tcp", address, true, &internal.Duration{Duration: time.Minute})
}

func TestOctetCountingStrictWithZeroKeepAlive_tcp_tls(t *testing.T) {
	testStrictOctetCounting(t, "tcp", address, true, &internal.Duration{Duration: 0})
}

func TestOctetCountingStrict_unix(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestStrict_unix.sock")
	testStrictOctetCounting(t, "unix", sock, false, nil)
}

func TestOctetCountingBestEffort_unix(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestBestEffort_unix.sock")
	testBestEffortOctetCounting(t, "unix", sock, false, nil)
}

func TestOctetCountingStrict_unix_tls(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestStrict_unix_tls.sock")
	testStrictOctetCounting(t, "unix", sock, true, nil)
}

func TestOctetCountingBestEffort_unix_tls(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestBestEffort_unix_tls.sock")
	testBestEffortOctetCounting(t, "unix", sock, true, nil)
}
