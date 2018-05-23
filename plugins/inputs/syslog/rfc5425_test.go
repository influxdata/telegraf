package syslog

import (
	"crypto/tls"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const (
	address = ":6514"
)

var (
	pki = testutil.NewPKI("../../../testutil/pki")
)

type testCase5425 struct {
	name           string
	data           []byte
	wantBestEffort []testutil.Metric
	wantStrict     []testutil.Metric
	werr           int // how many errors we expect in the strict mode?
}

func getTestCasesForRFC5425() []testCase5425 {
	testCases := []testCase5425{
		{
			name: "1st/avg/ok",
			data: []byte(`188 <29>1 2016-02-21T04:32:57+00:00 web1 someservice 2341 2 [origin][meta sequence="14125553" service="someservice"] "GET /v1/ok HTTP/1.1" 200 145 "-" "hacheck 0.9.0" 24306 127.0.0.1:40124 575`),
			wantStrict: []testutil.Metric{
				testutil.Metric{
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
			wantBestEffort: []testutil.Metric{
				testutil.Metric{
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
		},
		{
			name: "1st/min/ok//2nd/min/ok",
			data: []byte("16 <1>2 - - - - - -17 <4>11 - - - - - -"),
			wantStrict: []testutil.Metric{
				testutil.Metric{
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
				testutil.Metric{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version": uint16(11),
					},
					Tags: map[string]string{
						"severity":         "4",
						"severity_level":   "warning",
						"facility":         "0",
						"facility_message": "kernel messages",
					},
					Time: defaultTime,
				},
			},
			wantBestEffort: []testutil.Metric{
				testutil.Metric{
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
				testutil.Metric{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version": uint16(11),
					},
					Tags: map[string]string{
						"severity":         "4",
						"severity_level":   "warning",
						"facility":         "0",
						"facility_message": "kernel messages",
					},
					Time: defaultTime,
				},
			},
		},
		{
			name: "1st/utf8/ok",
			data: []byte("23 <1>1 - - - - - - hellø"),
			wantStrict: []testutil.Metric{
				testutil.Metric{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version": uint16(1),
						"message": "hellø",
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
			wantBestEffort: []testutil.Metric{
				testutil.Metric{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version": uint16(1),
						"message": "hellø",
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
		},
		{
			name: "1st/nl/ok", // newline
			data: []byte("28 <1>3 - - - - - - hello\nworld"),
			wantStrict: []testutil.Metric{
				testutil.Metric{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version": uint16(3),
						"message": "hello\nworld",
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
			wantBestEffort: []testutil.Metric{
				testutil.Metric{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version": uint16(3),
						"message": "hello\nworld",
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
		},
		{
			name:       "1st/uf/ko", // underflow (msglen less than provided octets)
			data:       []byte("16 <1>2"),
			wantStrict: nil,
			wantBestEffort: []testutil.Metric{
				testutil.Metric{
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
			},
			werr: 1,
		},
		{
			name: "1st/min/ok",
			data: []byte("16 <1>1 - - - - - -"),
			wantStrict: []testutil.Metric{
				testutil.Metric{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version": uint16(1),
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
			wantBestEffort: []testutil.Metric{
				testutil.Metric{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version": uint16(1),
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
		},
		{
			name:       "1st/uf/mf", // The first "underflow" message breaks also the second one
			data:       []byte("16 <1>217 <11>1 - - - - - -"),
			wantStrict: nil,
			wantBestEffort: []testutil.Metric{
				testutil.Metric{
					Measurement: "syslog",
					Fields: map[string]interface{}{
						"version": uint16(217),
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
				testutil.Metric{
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
			wantBestEffort: []testutil.Metric{
				testutil.Metric{
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
		},
	}

	return testCases
}

func newTCPSyslogReceiver(keepAlive *internal.Duration, maxConn int, bestEffort bool) *Syslog {
	d := &internal.Duration{
		Duration: defaultReadTimeout,
	}
	s := &Syslog{
		Protocol: "tcp",
		Address:  address,
		now: func() time.Time {
			return defaultTime
		},
		ReadTimeout: d,
		BestEffort:  bestEffort,
	}
	if keepAlive != nil {
		s.KeepAlivePeriod = keepAlive
	}
	if maxConn > 0 {
		s.MaxConnections = maxConn
	}

	return s
}

func testStrictRFC5425(t *testing.T, wantTLS bool, keepAlive *internal.Duration) {
	for _, tc := range getTestCasesForRFC5425() {
		t.Run(tc.name, func(t *testing.T) {
			// Creation of a strict mode receiver
			receiver := newTCPSyslogReceiver(keepAlive, 0, false)
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
				conn, err = tls.Dial("tcp", address, config)
			} else {
				conn, err = net.Dial("tcp", address)
				defer conn.Close()
			}
			require.NotNil(t, conn)
			require.NoError(t, err)

			// Clear
			acc.ClearMetrics()
			acc.Errors = make([]error, 0)

			// Write
			conn.Write(tc.data)

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

func testBestEffortRFC5425(t *testing.T, wantTLS bool, keepAlive *internal.Duration) {
	for _, tc := range getTestCasesForRFC5425() {
		t.Run(tc.name, func(t *testing.T) {
			// Creation of a best effort mode receiver
			receiver := newTCPSyslogReceiver(keepAlive, 0, true)
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
				conn, err = tls.Dial("tcp", address, config)
			} else {
				conn, err = net.Dial("tcp", address)
				defer conn.Close()
			}
			require.NotNil(t, conn)
			require.NoError(t, err)

			// Clear
			acc.ClearMetrics()
			acc.Errors = make([]error, 0)

			// Write
			conn.Write(tc.data)

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

func TestStrict_tcp(t *testing.T) {
	testStrictRFC5425(t, false, nil)
}

func TestBestEffort_tcp(t *testing.T) {
	testBestEffortRFC5425(t, false, nil)
}

func TestStrict_tcp_tls(t *testing.T) {
	testStrictRFC5425(t, true, nil)
}

func TestBestEffort_tcp_tls(t *testing.T) {
	testBestEffortRFC5425(t, true, nil)
}

func TestStrictWithKeepAlive_tcp_tls(t *testing.T) {
	testStrictRFC5425(t, true, &internal.Duration{Duration: time.Minute})
}

func TestStrictWithZeroKeepAlive_tcp_tls(t *testing.T) {
	testStrictRFC5425(t, true, &internal.Duration{Duration: 0})
}
