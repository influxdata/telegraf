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

func getTestCasesForNonTransparent(hasRemoteAddr bool) []testCaseStream {
	testCases := []testCaseStream{
		{
			name: "1st/avg/ok",
			data: []byte(
				`<29>1 2016-02-21T04:32:57+00:00 web1 someservice 2341 2 [origin][meta sequence="14125553" service="someservice"] ` +
					`"GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`,
			),
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

	if hasRemoteAddr {
		for _, tc := range testCases {
			for _, m := range tc.wantStrict {
				m.AddTag("source", "127.0.0.1")
			}
			for _, m := range tc.wantBestEffort {
				m.AddTag("source", "127.0.0.1")
			}
		}
	}

	return testCases
}

func testBestEffortNonTransparent(t *testing.T, protocol string, address string, wantTLS bool) {
	keepAlive := (*config.Duration)(nil)
	for _, tc := range getTestCasesForNonTransparent(protocol != "unix") {
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

			// Wait that the number of data points is accumulated
			// Since the receiver is running concurrently
			if tc.wantBestEffort != nil {
				acc.Wait(len(tc.wantBestEffort))
			}

			testutil.RequireMetricsEqual(t, tc.wantStrict, acc.GetTelegrafMetrics())
		})
	}
}

func TestNonTransparentBestEffort_tcp(t *testing.T) {
	testBestEffortNonTransparent(t, "tcp", address, false)
}

func TestNonTransparentBestEffort_tcp_tls(t *testing.T) {
	testBestEffortNonTransparent(t, "tcp", address, true)
}

func TestNonTransparentBestEffort_unix(t *testing.T) {
	sock := testutil.TempSocket(t)
	testBestEffortNonTransparent(t, "unix", sock, false)
}

func TestNonTransparentBestEffort_unix_tls(t *testing.T) {
	sock := testutil.TempSocket(t)
	testBestEffortNonTransparent(t, "unix", sock, true)
}
