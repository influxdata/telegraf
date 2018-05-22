package syslog

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const (
	serviceRootPEM = `-----BEGIN CERTIFICATE-----
MIIBxzCCATCgAwIBAgIJAJb7HqN2BzWWMA0GCSqGSIb3DQEBCwUAMBYxFDASBgNV
BAMMC1RlbGVncmFmIENBMB4XDTE3MTEwNDA0MzEwN1oXDTI3MTEwMjA0MzEwN1ow
FjEUMBIGA1UEAwwLVGVsZWdyYWYgQ0EwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJ
AoGBANbkUkK6JQC3rbLcXhLJTS9SX6uXyFwl7bUfpAN5Hm5EqfvG3PnLrogfTGLr
Tq5CRAu/gbbdcMoL9TLv/aaDVnrpV0FslKhqYmkOgT28bdmA7Qtr539aQpMKCfcW
WCnoMcBD5u5h9MsRqpdq+0Mjlsf1H2hSf07jHk5R1T4l8RMXAgMBAAGjHTAbMAwG
A1UdEwQFMAMBAf8wCwYDVR0PBAQDAgEGMA0GCSqGSIb3DQEBCwUAA4GBANSrwvpU
t8ihIhpHqgJZ34DM92CZZ3ZHmH/KyqlnuGzjjpnVZiXVrLDTOzrA0ziVhmefY29w
roHjENbFm54HW97ogxeURuO8HRHIVh2U0rkyVxOfGZiUdINHqsZdSnDY07bzCtSr
Z/KsfWXM5llD1Ig1FyBHpKjyUvfzr73sjm/4
-----END CERTIFICATE-----`
	serviceCertPEM = `-----BEGIN CERTIFICATE-----
MIIBzzCCATigAwIBAgIBATANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQDDAtUZWxl
Z3JhZiBDQTAeFw0xNzExMDQwNDMxMDdaFw0yNzExMDIwNDMxMDdaMBQxEjAQBgNV
BAMMCWxvY2FsaG9zdDCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEAsJRss1af
XKrcIjQoAp2kdJIpT2Ya+MRQXJ18b0PP7szh2lisY11kd/HCkd4D4efuIkpszHaN
xwyTOZLOoplxp6fizzgOYjXsJ6SzbO1MQNmq8Ch/+uKiGgFwLX+YxOOsGSDIHNhF
vcBi93cQtCWPBFz6QRQf9yfIAA5KKxUfJcMCAwEAAaMvMC0wCQYDVR0TBAIwADAL
BgNVHQ8EBAMCBSAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDQYJKoZIhvcNAQELBQAD
gYEAiC3WI4y9vfYz53gw7FKnNK7BBdwRc43x7Pd+5J/cclWyUZPdmcj1UNmv/3rj
2qcMmX06UdgPoHppzNAJePvMVk0vjMBUe9MmYlafMz0h4ma/it5iuldXwmejFcdL
6wWQp7gVTileCEmq9sNvfQN1FmT3EWf4IMdO2MNat/1If0g=
-----END CERTIFICATE-----`
	serviceKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQCwlGyzVp9cqtwiNCgCnaR0kilPZhr4xFBcnXxvQ8/uzOHaWKxj
XWR38cKR3gPh5+4iSmzMdo3HDJM5ks6imXGnp+LPOA5iNewnpLNs7UxA2arwKH/6
4qIaAXAtf5jE46wZIMgc2EW9wGL3dxC0JY8EXPpBFB/3J8gADkorFR8lwwIDAQAB
AoGBAJaFHxfMmjHK77U0UnrQWFSKFy64cftmlL4t/Nl3q7L68PdIKULWZIMeEWZ4
I0UZiFOwr4em83oejQ1ByGSwekEuiWaKUI85IaHfcbt+ogp9hY/XbOEo56OPQUAd
bEZv1JqJOqta9Ug1/E1P9LjEEyZ5F5ubx7813rxAE31qKtKJAkEA1zaMlCWIr+Rj
hGvzv5rlHH3wbOB4kQFXO4nqj3J/ttzR5QiJW24STMDcbNngFlVcDVju56LrNTiD
dPh9qvl7nwJBANILguR4u33OMksEZTYB7nQZSurqXsq6382zH7pTl29ANQTROHaM
PKC8dnDWq8RGTqKuvWblIzzGIKqIMovZo10CQC96T0UXirITFolOL3XjvAuvFO1Q
EAkdXJs77805m0dCK+P1IChVfiAEpBw3bKJArpAbQIlFfdI953JUp5SieU0CQEub
BSSEKMjh/cxu6peEHnb/262vayuCFKkQPu1sxWewLuVrAe36EKCy9dcsDmv5+rgo
Odjdxc9Madm4aKlaT6kCQQCpAgeblDrrxTrNQ+Typzo37PlnQrvI+0EceAUuJ72G
P0a+YZUeHNRqT2pPN9lMTAZGGi3CtcF2XScbLNEBeXge
-----END RSA PRIVATE KEY-----`
	clientRootPEM = serviceRootPEM
	clientCertPEM = `-----BEGIN CERTIFICATE-----
MIIBzjCCATegAwIBAgIBAjANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQDDAtUZWxl
Z3JhZiBDQTAeFw0xNzExMDQwNDMxMDdaFw0yNzExMDIwNDMxMDdaMBMxETAPBgNV
BAMMCHRlbGVncmFmMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDP2IMqyOqI
sJjwBprrz8WPzmlrpyYikQ4XSCSJB3DSTIO+igqMpBUTj3vLlOzsHfVVot1WRqc6
3esM4JE92rc6S73xi4g8L/r8cPIHW4hvFJdMti4UkJBWim8ArSbFqnZjcR19G3tG
LUOiXAUG3nWzMzoEsPruvV1dkKRbJVE4MwIDAQABoy8wLTAJBgNVHRMEAjAAMAsG
A1UdDwQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAjANBgkqhkiG9w0BAQsFAAOB
gQCHxMk38XNxL9nPFBYo3JqITJCFswu6/NLHwDBXCuZKl53rUuFWduiO+1OuScKQ
sQ79W0jHsWRKGOUFrF5/Gdnh8AlkVaITVlcmhdAOFCEbeGpeEvLuuK6grckPitxy
bRF5oM4TCLKKAha60Ir41rk2bomZM9+NZu+Bm+csDqCoxQ==
-----END CERTIFICATE-----`
	clientKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDP2IMqyOqIsJjwBprrz8WPzmlrpyYikQ4XSCSJB3DSTIO+igqM
pBUTj3vLlOzsHfVVot1WRqc63esM4JE92rc6S73xi4g8L/r8cPIHW4hvFJdMti4U
kJBWim8ArSbFqnZjcR19G3tGLUOiXAUG3nWzMzoEsPruvV1dkKRbJVE4MwIDAQAB
AoGAFzb/r4+xYoMXEfgq5ZvXXTCY5cVNpR6+jCsqqYODPnn9XRLeCsdo8z5bfWms
7NKLzHzca/6IPzL6Rf3vOxFq1YyIZfYVHH+d63/9blAm3Iajjp1W2yW5aj9BJjTb
nm6F0RfuW/SjrZ9IXxTZhSpCklPmUzVZpzvwV3KGeVTVCEECQQDoavCeOwLuqDpt
0aM9GMFUpOU7kLPDuicSwCDaTae4kN2rS17Zki41YXe8A8+509IEN7mK09Vq9HxY
SX6EmV1FAkEA5O9QcCHEa8P12EmUC8oqD2bjq6o7JjUIRlKinwZTlooMJYZw98gA
FVSngTUvLVCVIvSdjldXPOGgfYiccTZrFwJAfHS3gKOtAEuJbkEyHodhD4h1UB4+
hPLr9Xh4ny2yQH0ilpV3px5GLEOTMFUCKUoqTiPg8VxaDjn5U/WXED5n2QJAR4J1
NsFlcGACj+/TvacFYlA6N2nyFeokzoqLX28Ddxdh2erXqJ4hYIhT1ik9tkLggs2z
1T1084BquCuO6lIcOwJBALX4xChoMUF9k0IxSQzlz//seQYDkQNsE7y9IgAOXkzp
RaR4pzgPbnKj7atG+2dBnffWfE+1Mcy0INDAO6WxPg0=
-----END RSA PRIVATE KEY-----`
	address = ":6514"
)

var (
	initServiceCertFiles sync.Once
	serviceCAFile        string
	serviceCertFile      string
	serviceKeyFile       string
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

func newTLSSyslogReceiver(keepAlive *internal.Duration, maxConn int, bestEffort bool) *Syslog {
	initServiceCertFiles.Do(func() {
		scaf, err := ioutil.TempFile("", "serviceCAFile.crt")
		if err != nil {
			panic(err)
		}
		defer scaf.Close()
		_, err = io.Copy(scaf, bytes.NewReader([]byte(serviceRootPEM)))
		serviceCAFile = scaf.Name()

		scf, err := ioutil.TempFile("", "serviceCertFile.crt")
		if err != nil {
			panic(err)
		}
		defer scf.Close()
		_, err = io.Copy(scf, bytes.NewReader([]byte(serviceCertPEM)))
		serviceCertFile = scf.Name()

		skf, err := ioutil.TempFile("", "serviceKeyFile.crt")
		if err != nil {
			panic(err)
		}
		defer skf.Close()
		_, err = io.Copy(skf, bytes.NewReader([]byte(serviceKeyPEM)))
		serviceKeyFile = skf.Name()
	})

	receiver := newTCPSyslogReceiver(keepAlive, maxConn, bestEffort)
	receiver.Cacert = serviceCAFile
	receiver.Cert = serviceCertFile
	receiver.Key = serviceKeyFile

	return receiver
}

func getTLSSyslogSender() net.Conn {
	cas := x509.NewCertPool()
	cas.AppendCertsFromPEM([]byte(serviceRootPEM))
	clientCert, err := tls.X509KeyPair([]byte(clientCertPEM), []byte(clientKeyPEM))
	if err != nil {
		panic(err)
	}

	config := &tls.Config{
		RootCAs:            cas,
		Certificates:       []tls.Certificate{clientCert},
		MinVersion:         tls.VersionTLS12,
		MaxVersion:         tls.VersionTLS12,
		Renegotiation:      tls.RenegotiateNever,
		InsecureSkipVerify: false,
		ServerName:         "localhost",
	}

	c, err := tls.Dial("tcp", address, config)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	return c
}

func testStrict(t *testing.T, acc *testutil.Accumulator, tls bool) {
	for _, tc := range getTestCasesForRFC5425() {
		t.Run(tc.name, func(t *testing.T) {
			// Connect
			var conn net.Conn
			var err error
			if tls {
				conn = getTLSSyslogSender()

			} else {
				conn, err = net.Dial("tcp", address)
				defer conn.Close()
			}
			require.NotNil(t, conn)
			require.Nil(t, err)

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

func testBestEffort(t *testing.T, acc *testutil.Accumulator, tls bool) {
	for _, tc := range getTestCasesForRFC5425() {
		t.Run(tc.name, func(t *testing.T) {
			// Connect
			var conn net.Conn
			var err error
			if tls {
				conn = getTLSSyslogSender()
				require.NotNil(t, conn)
			} else {
				conn, err = net.Dial("tcp", address)
				require.NoError(t, err)
				defer conn.Close()
			}

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

func TestTCPInStrictMode(t *testing.T) {
	receiver := newTCPSyslogReceiver(nil, 0, false)

	acc := &testutil.Accumulator{}
	require.NoError(t, receiver.Start(acc))
	defer receiver.Stop()

	testStrict(t, acc, false)
}

func TestTCPInBestEffort(t *testing.T) {
	receiver := newTCPSyslogReceiver(nil, 0, true)

	acc := &testutil.Accumulator{}
	require.NoError(t, receiver.Start(acc))
	defer receiver.Stop()

	testBestEffort(t, acc, false)
}

func TestTLSInStrictMode(t *testing.T) {
	receiver := newTLSSyslogReceiver(nil, 0, false)

	acc := &testutil.Accumulator{}
	require.NoError(t, receiver.Start(acc))
	defer receiver.Stop()

	testStrict(t, acc, true)
}

func TestTLSInBestEffortOn(t *testing.T) {
	receiver := newTLSSyslogReceiver(nil, 0, true)
	require.True(t, receiver.BestEffort)

	acc := &testutil.Accumulator{}
	require.NoError(t, receiver.Start(acc))
	defer receiver.Stop()

	testBestEffort(t, acc, true)
}

func TestTLSWithKeepAliveInStrictMode(t *testing.T) {
	keepAlivePeriod := &internal.Duration{
		Duration: time.Minute,
	}
	receiver := newTLSSyslogReceiver(keepAlivePeriod, 0, false)
	require.Equal(t, receiver.KeepAlivePeriod, keepAlivePeriod)

	acc := &testutil.Accumulator{}
	require.NoError(t, receiver.Start(acc))
	defer receiver.Stop()

	testStrict(t, acc, true)
}

func TestTLSWithZeroKeepAliveInStrictMode(t *testing.T) {
	keepAlivePeriod := &internal.Duration{
		Duration: 0,
	}
	receiver := newTLSSyslogReceiver(keepAlivePeriod, 0, false)
	require.Equal(t, receiver.KeepAlivePeriod, keepAlivePeriod)

	acc := &testutil.Accumulator{}
	require.NoError(t, receiver.Start(acc))
	defer receiver.Stop()

	testStrict(t, acc, true)
}
