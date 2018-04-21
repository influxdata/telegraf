package http_listener

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf/testutil"

	"github.com/stretchr/testify/require"
)

const (
	testMsg = "cpu_load_short,host=server01 value=12.0 1422568543702900257\n"

	testMsgNoNewline = "cpu_load_short,host=server01 value=12.0 1422568543702900257"

	testMsgs = `cpu_load_short,host=server02 value=12.0 1422568543702900257
cpu_load_short,host=server03 value=12.0 1422568543702900257
cpu_load_short,host=server04 value=12.0 1422568543702900257
cpu_load_short,host=server05 value=12.0 1422568543702900257
cpu_load_short,host=server06 value=12.0 1422568543702900257
`
	badMsg = "blahblahblah: 42\n"

	emptyMsg = ""

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

	basicUsername = "test-username-please-ignore"
	basicPassword = "super-secure-password!"
)

var (
	initClient           sync.Once
	client               *http.Client
	initServiceCertFiles sync.Once
	allowedCAFiles       []string
	serviceCAFiles       []string
	serviceCertFile      string
	serviceKeyFile       string
)

func newTestHTTPListener() *HTTPListener {
	listener := &HTTPListener{
		ServiceAddress: "localhost:0",
		TimeFunc:       time.Now,
	}
	return listener
}

func newTestHTTPAuthListener() *HTTPListener {
	listener := newTestHTTPListener()
	listener.BasicUsername = basicUsername
	listener.BasicPassword = basicPassword
	return listener
}

func newTestHTTPSListener() *HTTPListener {
	initServiceCertFiles.Do(func() {
		acaf, err := ioutil.TempFile("", "allowedCAFile.crt")
		if err != nil {
			panic(err)
		}
		defer acaf.Close()
		_, err = io.Copy(acaf, bytes.NewReader([]byte(clientRootPEM)))
		allowedCAFiles = []string{acaf.Name()}

		scaf, err := ioutil.TempFile("", "serviceCAFile.crt")
		if err != nil {
			panic(err)
		}
		defer scaf.Close()
		_, err = io.Copy(scaf, bytes.NewReader([]byte(serviceRootPEM)))
		serviceCAFiles = []string{scaf.Name()}

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

	listener := &HTTPListener{
		ServiceAddress:    "localhost:0",
		TlsAllowedCacerts: allowedCAFiles,
		TlsCert:           serviceCertFile,
		TlsKey:            serviceKeyFile,
		TimeFunc:          time.Now,
	}

	return listener
}

func getHTTPSClient() *http.Client {
	initClient.Do(func() {
		cas := x509.NewCertPool()
		cas.AppendCertsFromPEM([]byte(serviceRootPEM))
		clientCert, err := tls.X509KeyPair([]byte(clientCertPEM), []byte(clientKeyPEM))
		if err != nil {
			panic(err)
		}
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:            cas,
					Certificates:       []tls.Certificate{clientCert},
					MinVersion:         tls.VersionTLS12,
					MaxVersion:         tls.VersionTLS12,
					CipherSuites:       []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
					Renegotiation:      tls.RenegotiateNever,
					InsecureSkipVerify: false,
				},
			},
		}
	})
	return client
}

func createURL(listener *HTTPListener, scheme string, path string, rawquery string) string {
	u := url.URL{
		Scheme:   scheme,
		Host:     "localhost:" + strconv.Itoa(listener.Port),
		Path:     path,
		RawQuery: rawquery,
	}
	return u.String()
}

func TestWriteHTTPSNoClientAuth(t *testing.T) {
	listener := newTestHTTPSListener()
	listener.TlsAllowedCacerts = nil

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	cas := x509.NewCertPool()
	cas.AppendCertsFromPEM([]byte(serviceRootPEM))
	noClientAuthClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: cas,
			},
		},
	}

	// post single message to listener
	resp, err := noClientAuthClient.Post(createURL(listener, "https", "/write", "db=mydb"), "", bytes.NewBuffer([]byte(testMsg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)
}

func TestWriteHTTPSWithClientAuth(t *testing.T) {
	listener := newTestHTTPSListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := getHTTPSClient().Post(createURL(listener, "https", "/write", "db=mydb"), "", bytes.NewBuffer([]byte(testMsg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)
}

func TestWriteHTTPBasicAuth(t *testing.T) {
	listener := newTestHTTPAuthListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	client := &http.Client{}

	req, err := http.NewRequest("POST", createURL(listener, "http", "/write", "db=mydb"), bytes.NewBuffer([]byte(testMsg)))
	require.NoError(t, err)
	req.SetBasicAuth(basicUsername, basicPassword)
	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, http.StatusNoContent, resp.StatusCode)
}

func TestWriteHTTP(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := http.Post(createURL(listener, "http", "/write", "db=mydb"), "", bytes.NewBuffer([]byte(testMsg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)

	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(12)},
		map[string]string{"host": "server01"},
	)

	// post multiple message to listener
	resp, err = http.Post(createURL(listener, "http", "/write", "db=mydb"), "", bytes.NewBuffer([]byte(testMsgs)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)

	acc.Wait(2)
	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "cpu_load_short",
			map[string]interface{}{"value": float64(12)},
			map[string]string{"host": hostTag},
		)
	}

	// Post a gigantic metric to the listener and verify that an error is returned:
	resp, err = http.Post(createURL(listener, "http", "/write", "db=mydb"), "", bytes.NewBuffer([]byte(hugeMetric)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 400, resp.StatusCode)

	acc.Wait(3)
	acc.AssertContainsTaggedFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(12)},
		map[string]string{"host": "server01"},
	)
}

// http listener should add a newline at the end of the buffer if it's not there
func TestWriteHTTPNoNewline(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := http.Post(createURL(listener, "http", "/write", "db=mydb"), "", bytes.NewBuffer([]byte(testMsgNoNewline)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)

	acc.Wait(1)
	acc.AssertContainsTaggedFields(t, "cpu_load_short",
		map[string]interface{}{"value": float64(12)},
		map[string]string{"host": "server01"},
	)
}

func TestWriteHTTPMaxLineSizeIncrease(t *testing.T) {
	listener := &HTTPListener{
		ServiceAddress: "localhost:0",
		MaxLineSize:    128 * 1000,
		TimeFunc:       time.Now,
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// Post a gigantic metric to the listener and verify that it writes OK this time:
	resp, err := http.Post(createURL(listener, "http", "/write", "db=mydb"), "", bytes.NewBuffer([]byte(hugeMetric)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)
}

func TestWriteHTTPVerySmallMaxBody(t *testing.T) {
	listener := &HTTPListener{
		ServiceAddress: "localhost:0",
		MaxBodySize:    4096,
		TimeFunc:       time.Now,
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(hugeMetric)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 413, resp.StatusCode)
}

func TestWriteHTTPVerySmallMaxLineSize(t *testing.T) {
	listener := &HTTPListener{
		ServiceAddress: "localhost:0",
		MaxLineSize:    70,
		TimeFunc:       time.Now,
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(testMsgs)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)

	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	acc.Wait(len(hostTags))
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "cpu_load_short",
			map[string]interface{}{"value": float64(12)},
			map[string]string{"host": hostTag},
		)
	}
}

func TestWriteHTTPLargeLinesSkipped(t *testing.T) {
	listener := &HTTPListener{
		ServiceAddress: "localhost:0",
		MaxLineSize:    100,
		TimeFunc:       time.Now,
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	resp, err := http.Post(createURL(listener, "http", "/write", ""), "", bytes.NewBuffer([]byte(hugeMetric+testMsgs)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 400, resp.StatusCode)

	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	acc.Wait(len(hostTags))
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "cpu_load_short",
			map[string]interface{}{"value": float64(12)},
			map[string]string{"host": hostTag},
		)
	}
}

// test that writing gzipped data works
func TestWriteHTTPGzippedData(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	data, err := ioutil.ReadFile("./testdata/testmsgs.gz")
	require.NoError(t, err)

	req, err := http.NewRequest("POST", createURL(listener, "http", "/write", ""), bytes.NewBuffer(data))
	require.NoError(t, err)
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.EqualValues(t, 204, resp.StatusCode)

	hostTags := []string{"server02", "server03",
		"server04", "server05", "server06"}
	acc.Wait(len(hostTags))
	for _, hostTag := range hostTags {
		acc.AssertContainsTaggedFields(t, "cpu_load_short",
			map[string]interface{}{"value": float64(12)},
			map[string]string{"host": hostTag},
		)
	}
}

// writes 25,000 metrics to the listener with 10 different writers
func TestWriteHTTPHighTraffic(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping due to hang on darwin")
	}
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post many messages to listener
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(innerwg *sync.WaitGroup) {
			defer innerwg.Done()
			for i := 0; i < 500; i++ {
				resp, err := http.Post(createURL(listener, "http", "/write", "db=mydb"), "", bytes.NewBuffer([]byte(testMsgs)))
				require.NoError(t, err)
				resp.Body.Close()
				require.EqualValues(t, 204, resp.StatusCode)
			}
		}(&wg)
	}

	wg.Wait()
	listener.Gather(acc)

	acc.Wait(25000)
	require.Equal(t, int64(25000), int64(acc.NMetrics()))
}

func TestReceive404ForInvalidEndpoint(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := http.Post(createURL(listener, "http", "/foobar", ""), "", bytes.NewBuffer([]byte(testMsg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 404, resp.StatusCode)
}

func TestWriteHTTPInvalid(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := http.Post(createURL(listener, "http", "/write", "db=mydb"), "", bytes.NewBuffer([]byte(badMsg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 400, resp.StatusCode)
}

func TestWriteHTTPEmpty(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post single message to listener
	resp, err := http.Post(createURL(listener, "http", "/write", "db=mydb"), "", bytes.NewBuffer([]byte(emptyMsg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)
}

func TestQueryAndPingHTTP(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	// post query to listener
	resp, err := http.Post(
		createURL(listener, "http", "/query", "db=&q=CREATE+DATABASE+IF+NOT+EXISTS+%22mydb%22"), "", nil)
	require.NoError(t, err)
	require.EqualValues(t, 200, resp.StatusCode)

	// post ping to listener
	resp, err = http.Post(createURL(listener, "http", "/ping", ""), "", nil)
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)
}

func TestWriteWithPrecision(t *testing.T) {
	listener := newTestHTTPListener()

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	msg := "xyzzy value=42 1422568543\n"
	resp, err := http.Post(
		createURL(listener, "http", "/write", "precision=s"), "", bytes.NewBuffer([]byte(msg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)

	acc.Wait(1)
	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, time.Unix(0, 1422568543000000000), acc.Metrics[0].Time)
}

func TestWriteWithPrecisionNoTimestamp(t *testing.T) {
	listener := newTestHTTPListener()
	listener.TimeFunc = func() time.Time {
		return time.Unix(42, 123456789)
	}

	acc := &testutil.Accumulator{}
	require.NoError(t, listener.Start(acc))
	defer listener.Stop()

	msg := "xyzzy value=42\n"
	resp, err := http.Post(
		createURL(listener, "http", "/write", "precision=s"), "", bytes.NewBuffer([]byte(msg)))
	require.NoError(t, err)
	resp.Body.Close()
	require.EqualValues(t, 204, resp.StatusCode)

	acc.Wait(1)
	require.Equal(t, 1, len(acc.Metrics))
	require.Equal(t, time.Unix(42, 0), acc.Metrics[0].Time)
}

const hugeMetric = `super_long_metric,foo=bar clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i
`
