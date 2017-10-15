package http_listener

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
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
MIIBxzCCATCgAwIBAgIJAOLq2g9+9TVgMA0GCSqGSIb3DQEBCwUAMBYxFDASBgNV
BAMMC1RlbGVncmFmIENBMB4XDTE3MTAwMjIyNDMwOFoXDTE3MTEwMTIyNDMwOFow
FjEUMBIGA1UEAwwLVGVsZWdyYWYgQ0EwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJ
AoGBALHtGXLKZz3HUA4E1H0mR3gAtgNwUSRArxylCjQwO/7tFEYDFVCCPFzAF7G8
hzHyBNgx5FwNrH3bMEol9iIxzoZNU0XTWS7DzN4S+89C2Tn+NaFko/SeFBMp4IK/
55YAgcYGe2QbFnPITGYPT05VkbSBMD0PBITNSwsclGZGFVoHAgMBAAGjHTAbMAwG
A1UdEwQFMAMBAf8wCwYDVR0PBAQDAgEGMA0GCSqGSIb3DQEBCwUAA4GBAIJpAA+X
QB57JhNxevUlFFLmGx7ASKrOeZLupzak4qUK718erafMAsXhydx1eKL/5Ne7ZcFa
Tf6dRPzCjv89WzYK/kJ59AgATkXNPvADRUKd0ViQw4Q4EcfuQrTMEym+gl1W2qQl
U9/eBDE341pcrfdHHGhS5LKv6KTmjyYmDLxl
-----END CERTIFICATE-----`
	serviceCertPEM = `-----BEGIN CERTIFICATE-----
MIIBzzCCATigAwIBAgIBATANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQDDAtUZWxl
Z3JhZiBDQTAeFw0xNzEwMDIyMjQzMDhaFw0yNzA5MzAyMjQzMDhaMBQxEjAQBgNV
BAMMCWxvY2FsaG9zdDCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEAoI/8ceps
DvvA3KUDViYwZcB+RvfT6XCPCT35mEzuXWP42JHk1VPNA41215U8CGoJF7+OzRcZ
an3b2WLfAph+bi4Vmpe8eolmPHjf57jJ2fdDeLtMA4T0WF8yR4fHxrrU2UFsgXod
kpQNqa/R5+iEKNMQVQgD2HjP5BE1u+H6fscCAwEAAaMvMC0wCQYDVR0TBAIwADAL
BgNVHQ8EBAMCBSAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDQYJKoZIhvcNAQELBQAD
gYEAV5vx8FHNlD6Z3e01/MiNSmYXn93CwlixYMRyW1Ri2P6hMtJiMRp59fNFzroa
iv6djr30uuKYOiAvdKhNaYWERgrtjGVEuPoIMQfaAaKHQj6CKLBXeGZ5Gxhy+M6G
OE6g0E4ufHOqr1h1GDIiAq88zyJ2AupgLLUCMFtkq0v0mr0=
-----END CERTIFICATE-----`
	serviceKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCgj/xx6mwO+8DcpQNWJjBlwH5G99PpcI8JPfmYTO5dY/jYkeTV
U80DjXbXlTwIagkXv47NFxlqfdvZYt8CmH5uLhWal7x6iWY8eN/nuMnZ90N4u0wD
hPRYXzJHh8fGutTZQWyBeh2SlA2pr9Hn6IQo0xBVCAPYeM/kETW74fp+xwIDAQAB
AoGABiRb6NOp3Ize3NHnJcWCNnI9omNalOR8ZEMdqCjROXtYiphSI6L4BbnEoQyR
ZlUAEgt+3/ORQlScM12n4EaLF4Zi4CTGmibRHUff/ybUDGMg2Lp/AL/ghP/3U37l
C/oRjohK9Rqn28hf8xgL9Jz+KbQaVv5f+frLwL3EKreYtOkCQQDLe1s89rbxvTZr
PhtwYrnXC8KbBNPIzJbTXrphqr0H3xuDlTpd+4tvIlL6LoqANYXAmHHlKUuPcar6
QCj9xNwTAkEAygDRac8qewqIWhZOs0u8phC37dxzwVXslrgjO+kTLxN/Q1srK45T
gHDbJuCrBPkYrjAXWHd2rIkOWl0rk38A/QJADct4HQLw1iSous6EF7Npu+19LPs/
zF4qX3wNkK99jzoN6HbGdTandkpSa8mZ9CUswyjSl+Gb0Ma4+6w72zBsZwJBAKn+
Zj0VCjrhcj3d5/0bD3bxOtgBXaimFqP/8ibIzkwfrEmSv5G4BK1iTAs7prBYsFxm
PD9GyagI7vs8zR8jEkECQD51jhM8DDPah/ECC31we54Y9dqBOupy1a8y6os1YFkv
BV7zTVrpOzwUsrkMW+wFyQSX9eyyMfJHJihlobXA+QY=
-----END RSA PRIVATE KEY-----`
	clientRootPEM = `-----BEGIN CERTIFICATE-----
MIIBxzCCATCgAwIBAgIJAOLq2g9+9TVgMA0GCSqGSIb3DQEBCwUAMBYxFDASBgNV
BAMMC1RlbGVncmFmIENBMB4XDTE3MTAwMjIyNDMwOFoXDTE3MTEwMTIyNDMwOFow
FjEUMBIGA1UEAwwLVGVsZWdyYWYgQ0EwgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJ
AoGBALHtGXLKZz3HUA4E1H0mR3gAtgNwUSRArxylCjQwO/7tFEYDFVCCPFzAF7G8
hzHyBNgx5FwNrH3bMEol9iIxzoZNU0XTWS7DzN4S+89C2Tn+NaFko/SeFBMp4IK/
55YAgcYGe2QbFnPITGYPT05VkbSBMD0PBITNSwsclGZGFVoHAgMBAAGjHTAbMAwG
A1UdEwQFMAMBAf8wCwYDVR0PBAQDAgEGMA0GCSqGSIb3DQEBCwUAA4GBAIJpAA+X
QB57JhNxevUlFFLmGx7ASKrOeZLupzak4qUK718erafMAsXhydx1eKL/5Ne7ZcFa
Tf6dRPzCjv89WzYK/kJ59AgATkXNPvADRUKd0ViQw4Q4EcfuQrTMEym+gl1W2qQl
U9/eBDE341pcrfdHHGhS5LKv6KTmjyYmDLxl
-----END CERTIFICATE-----`
	clientCertPEM = `-----BEGIN CERTIFICATE-----
MIIBzjCCATegAwIBAgIBAjANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQDDAtUZWxl
Z3JhZiBDQTAeFw0xNzEwMDIyMjQzMDhaFw0yNzA5MzAyMjQzMDhaMBMxETAPBgNV
BAMMCHRlbGVncmFmMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDIrPGv8Sm1
6tI+vlATzWGOK1D40iNTiGj4FpcS2Tm4SdaDSfa3VL9N5l8aeuN4E8O2YXK3QcR8
NoeY87cWW06PtFc/ByS42VeWDKt28/DpGzbrzCVNOumS3X5QEyySYLpi0uqI9ZZ5
O2sOJ2yVua8F3cwqPTveVmU3LeQfVrh7QwIDAQABoy8wLTAJBgNVHRMEAjAAMAsG
A1UdDwQEAwIHgDATBgNVHSUEDDAKBggrBgEFBQcDAjANBgkqhkiG9w0BAQsFAAOB
gQAVEfHePY9fumW8rkbbSbiuQ1dGIINbMGPO17eAjOxMT4Z1jDb8oTVHbaZM0rKo
wKx4dDp5mnLK+NuMZ1sNxKOf6IMmQ022ANOYM0dkwfg13bpC3BGW8Z7nOFK0xXh6
4KTcXktBUtubmn6w7szvWY2OajPVoiGgcapwwhCrBEa6rg==
-----END CERTIFICATE-----`
	clientKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDIrPGv8Sm16tI+vlATzWGOK1D40iNTiGj4FpcS2Tm4SdaDSfa3
VL9N5l8aeuN4E8O2YXK3QcR8NoeY87cWW06PtFc/ByS42VeWDKt28/DpGzbrzCVN
OumS3X5QEyySYLpi0uqI9ZZ5O2sOJ2yVua8F3cwqPTveVmU3LeQfVrh7QwIDAQAB
AoGAHtvpdqLhRSZNGnTtn33vyIsEsp6t7ASID855gN6Cr8I7CIlxNRQFLxeD/HB1
VlvDtuIZX/DvJCLGi1C/EOMNm2nY7IT2gZgMpxvmfjfGhHKT1MWYu9cdyiOOacqD
yRDAcKpubIPEIV3aczglv9sVApXwZcgePzDwweTVfP/Nv5ECQQDthIv5Y5k3UO8h
Hht+27W8McFJ5eiF5OcLGOQ4nKGHkCOskfD4u/i+j+4dUeGBdpT8CzszgofBa6wh
dJevQerVAkEA2Ep8PUfXRjel8NiLNL8iK/SR26y8wPduKam31SMUPq71+GROKkFz
yYYAbKORs+fS6LBT+M48cEu470o+g8eptwJBALzCEMeSOqp2XIRSPAG2NBiq5fSH
jSIThvYPwxemisyEZYV4uivCnu06zz5n2zIa/k3L0zGdc6vomPRBh2aVmT0CQQCY
/B5ibfUbqnLKJzBXb7Xo50Vf3w9nYdvexjfMHtLL/47lUXVkOAWBDjIwpYWCfb/V
bBsJCj7/ot+9CYOsTEaDAkEA4XAGFxx78JMVuJLjevkf0pGUPEocdoOAvpYWT5sR
9FODrPEtW84ZevSmuByjzeqVzS3ElIxACopRJgSN20d9vg==
-----END RSA PRIVATE KEY-----`
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
		ServiceAddress: ":0",
	}
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
		ServiceAddress:    ":0",
		TlsAllowedCacerts: allowedCAFiles,
		TlsCert:           serviceCertFile,
		TlsKey:            serviceKeyFile,
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
		ServiceAddress: ":0",
		MaxLineSize:    128 * 1000,
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
		ServiceAddress: ":0",
		MaxBodySize:    4096,
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
		ServiceAddress: ":0",
		MaxLineSize:    70,
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
		ServiceAddress: ":0",
		MaxLineSize:    100,
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

const hugeMetric = `super_long_metric,foo=bar clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i,clients=1i,connected_slaves=0i,evicted_keys=0i,expired_keys=0i,instantaneous_ops_per_sec=0i,keyspace_hitrate=0,keyspace_hits=0i,keyspace_misses=2i,latest_fork_usec=0i,master_repl_offset=0i,mem_fragmentation_ratio=3.58,pubsub_channels=0i,pubsub_patterns=0i,rdb_changes_since_last_save=0i,repl_backlog_active=0i,repl_backlog_histlen=0i,repl_backlog_size=1048576i,sync_full=0i,sync_partial_err=0i,sync_partial_ok=0i,total_commands_processed=4i,total_connections_received=2i,uptime=869i,used_cpu_sys=0.07,used_cpu_sys_children=0,used_cpu_user=0.1,used_cpu_user_children=0,used_memory=502048i,used_memory_lua=33792i,used_memory_peak=501128i,used_memory_rss=1798144i
`
