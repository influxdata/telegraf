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

func getTestCasesForRFC6587() []testCaseStream {
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
		},
	}
	return testCases
}

func testStrictRFC6587(t *testing.T, protocol string, address string, wantTLS bool, keepAlive *internal.Duration) {
	for _, tc := range getTestCasesForRFC6587() {
		t.Run(tc.name, func(t *testing.T) {
			// Creation of a strict mode receiver
			receiver := newTCPSyslogReceiver(protocol+"://"+address, keepAlive, 0, false, false)
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

func testBestEffortRFC6587(t *testing.T, protocol string, address string, wantTLS bool, keepAlive *internal.Duration) {
	for _, tc := range getTestCasesForRFC6587() {
		t.Run(tc.name, func(t *testing.T) {
			// Creation of a best effort mode receiver
			receiver := newTCPSyslogReceiver(protocol+"://"+address, keepAlive, 0, true, false)
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

func TestRFC6587Strict_tcp(t *testing.T) {
	testStrictRFC6587(t, "tcp", address, false, nil)
}

func TestRFC6587BestEffort_tcp(t *testing.T) {
	testBestEffortRFC6587(t, "tcp", address, false, nil)
}

func TestRFC6587Strict_tcp_tls(t *testing.T) {
	testStrictRFC6587(t, "tcp", address, true, nil)
}

func TestRFC6587BestEffort_tcp_tls(t *testing.T) {
	testBestEffortRFC6587(t, "tcp", address, true, nil)
}

func TestRFC6587StrictWithKeepAlive_tcp_tls(t *testing.T) {
	testStrictRFC6587(t, "tcp", address, true, &internal.Duration{Duration: time.Minute})
}

func TestRFC6587StrictWithZeroKeepAlive_tcp_tls(t *testing.T) {
	testStrictRFC6587(t, "tcp", address, true, &internal.Duration{Duration: 0})
}

func TestRFC6587Strict_unix(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestStrict_unix.sock")
	testStrictRFC6587(t, "unix", sock, false, nil)
}

func TestRFC6587BestEffort_unix(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestBestEffort_unix.sock")
	testBestEffortRFC6587(t, "unix", sock, false, nil)
}

func TestRFC6587Strict_unix_tls(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestStrict_unix_tls.sock")
	testStrictRFC6587(t, "unix", sock, true, nil)
}

func TestRFC6587BestEffort_unix_tls(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "telegraf")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	sock := filepath.Join(tmpdir, "syslog.TestBestEffort_unix_tls.sock")
	testBestEffortRFC6587(t, "unix", sock, true, nil)
}
