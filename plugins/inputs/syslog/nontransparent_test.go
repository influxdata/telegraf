package syslog

import (
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	framing "github.com/influxdata/telegraf/internal/syslog"
	"github.com/influxdata/telegraf/testutil"
)

func getTestCasesForNonTransparent() []testCaseStream {
	testCases := []testCaseStream{
		{
			name: "1st/avg/ok",
			data: []byte(`<29>1 2016-02-21T04:32:57+00:00 web1 someservice 2341 2 [origin][meta sequence="14125553" service="someservice"] "GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`),
			wantStrict: []telegraf.Metric{
				testutil.MustMetric(
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
			wantBestEffort: []telegraf.Metric{
				testutil.MustMetric(
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
			werr: 1,
		},
		{
			name: "1st/min/ok//2nd/min/ok",
			data: []byte("<1>2 - - - - - -\n<4>11 - - - - - -\n"),
			wantStrict: []telegraf.Metric{
				testutil.MustMetric(
					"syslog",
					map[string]string{
						"severity": "alert",
						"facility": "kern",
					},
					map[string]interface{}{
						"version":       uint16(2),
						"severity_code": 1,
						"facility_code": 0,
					},
					defaultTime,
				),
				testutil.MustMetric(
					"syslog",
					map[string]string{
						"severity": "warning",
						"facility": "kern",
					},
					map[string]interface{}{
						"version":       uint16(11),
						"severity_code": 4,
						"facility_code": 0,
					},
					defaultTime.Add(time.Nanosecond),
				),
			},
			wantBestEffort: []telegraf.Metric{
				testutil.MustMetric(
					"syslog",
					map[string]string{
						"severity": "alert",
						"facility": "kern",
					},
					map[string]interface{}{
						"version":       uint16(2),
						"severity_code": 1,
						"facility_code": 0,
					},
					defaultTime,
				),
				testutil.MustMetric(
					"syslog",
					map[string]string{
						"severity": "warning",
						"facility": "kern",
					},
					map[string]interface{}{
						"version":       uint16(11),
						"severity_code": 4,
						"facility_code": 0,
					},
					defaultTime.Add(time.Nanosecond),
				),
			},
		},
	}
	return testCases
}

func testStrictNonTransparent(t *testing.T, protocol string, address string, wantTLS bool, keepAlive *config.Duration) {
	for _, tc := range getTestCasesForNonTransparent() {
		t.Run(tc.name, func(t *testing.T) {
			// Creation of a strict mode receiver
			receiver := newTCPSyslogReceiver(protocol+"://"+address, keepAlive, 10, false, framing.NonTransparent)
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
				require.NotNil(t, conn)
				require.NoError(t, err)
			} else {
				conn, err = net.Dial(protocol, address)
				require.NotNil(t, conn)
				require.NoError(t, err)
				defer conn.Close()
			}

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
			testutil.RequireMetricsEqual(t, tc.wantStrict, acc.GetTelegrafMetrics())
		})
	}
}

func testBestEffortNonTransparent(t *testing.T, protocol string, address string, wantTLS bool) {
	keepAlive := (*config.Duration)(nil)
	for _, tc := range getTestCasesForNonTransparent() {
		t.Run(tc.name, func(t *testing.T) {
			// Creation of a best effort mode receiver
			receiver := newTCPSyslogReceiver(protocol+"://"+address, keepAlive, 10, true, framing.NonTransparent)
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

			testutil.RequireMetricsEqual(t, tc.wantStrict, acc.GetTelegrafMetrics())
		})
	}
}

func TestNonTransparentStrict_tcp(t *testing.T) {
	testStrictNonTransparent(t, "tcp", address, false, nil)
}

func TestNonTransparentBestEffort_tcp(t *testing.T) {
	testBestEffortNonTransparent(t, "tcp", address, false)
}

func TestNonTransparentStrict_tcp_tls(t *testing.T) {
	testStrictNonTransparent(t, "tcp", address, true, nil)
}

func TestNonTransparentBestEffort_tcp_tls(t *testing.T) {
	testBestEffortNonTransparent(t, "tcp", address, true)
}

func TestNonTransparentStrictWithKeepAlive_tcp_tls(t *testing.T) {
	d := config.Duration(time.Minute)
	testStrictNonTransparent(t, "tcp", address, true, &d)
}

func TestNonTransparentStrictWithZeroKeepAlive_tcp_tls(t *testing.T) {
	d := config.Duration(0)
	testStrictNonTransparent(t, "tcp", address, true, &d)
}

func TestNonTransparentStrict_unix(t *testing.T) {
	sock := testutil.TempSocket(t)
	testStrictNonTransparent(t, "unix", sock, false, nil)
}

func TestNonTransparentBestEffort_unix(t *testing.T) {
	sock := testutil.TempSocket(t)
	testBestEffortNonTransparent(t, "unix", sock, false)
}

func TestNonTransparentStrict_unix_tls(t *testing.T) {
	sock := testutil.TempSocket(t)
	testStrictNonTransparent(t, "unix", sock, true, nil)
}

func TestNonTransparentBestEffort_unix_tls(t *testing.T) {
	sock := testutil.TempSocket(t)
	testBestEffortNonTransparent(t, "unix", sock, true)
}
