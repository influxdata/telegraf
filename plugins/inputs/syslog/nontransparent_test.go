package syslog

import (
	"crypto/tls"
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

func getTestCasesForNonTransparent() []testCaseStream {
	testCases := []testCaseStream{
		{
			name: "1st/avg/ok",
			data: []byte(`<29>1 2016-02-21T04:32:57+00:00 web1 someservice 2341 2 [origin][meta sequence="14125553" service="someservice"] "GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`),
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
			werr: 1,
		},
		{
			name: "1st/min/ok//2nd/min/ok",
			data: []byte("<1>2 - - - - - -\n<4>11 - - - - - -\n"),
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
	}
	return testCases
}

func testStrictNonTransparent(t *testing.T, protocol string, address string, wantTLS bool, keepAlive *internal.Duration) {
	for _, tc := range getTestCasesForNonTransparent() {
		t.Run(tc.name, func(t *testing.T) {
			// Creation of a strict mode receiver
			receiver := newTCPSyslogReceiver(protocol+"://"+address, keepAlive, 0, false, NonTransparent)
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

func testBestEffortNonTransparent(t *testing.T, protocol string, address string, wantTLS bool, keepAlive *internal.Duration) {
	for _, tc := range getTestCasesForNonTransparent() {
		t.Run(tc.name, func(t *testing.T) {
			// Creation of a best effort mode receiver
			receiver := newTCPSyslogReceiver(protocol+"://"+address, keepAlive, 0, true, NonTransparent)
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

func TestNonTransparentStrict_tcp(t *testing.T) {
	testStrictNonTransparent(t, "tcp", address, false, nil)
}

func TestNonTransparentBestEffort_tcp(t *testing.T) {
	testBestEffortNonTransparent(t, "tcp", address, false, nil)
}

func TestNonTransparentStrict_tcp_tls(t *testing.T) {
	testStrictNonTransparent(t, "tcp", address, true, nil)
}

func TestNonTransparentBestEffort_tcp_tls(t *testing.T) {
	testBestEffortNonTransparent(t, "tcp", address, true, nil)
}

func TestNonTransparentStrictWithKeepAlive_tcp_tls(t *testing.T) {
	testStrictNonTransparent(t, "tcp", address, true, &internal.Duration{Duration: time.Minute})
}

func TestNonTransparentStrictWithZeroKeepAlive_tcp_tls(t *testing.T) {
	testStrictNonTransparent(t, "tcp", address, true, &internal.Duration{Duration: 0})
}

func TestNonTransparentStrict_unix(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestStrict_unix.sock")
	testStrictNonTransparent(t, "unix", sock, false, nil)
}

func TestNonTransparentBestEffort_unix(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestBestEffort_unix.sock")
	testBestEffortNonTransparent(t, "unix", sock, false, nil)
}

func TestNonTransparentStrict_unix_tls(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestStrict_unix_tls.sock")
	testStrictNonTransparent(t, "unix", sock, true, nil)
}

func TestNonTransparentBestEffort_unix_tls(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestBestEffort_unix_tls.sock")
	testBestEffortNonTransparent(t, "unix", sock, true, nil)
}
