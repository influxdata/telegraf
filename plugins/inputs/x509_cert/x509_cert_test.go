package x509_cert

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	_tls "github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
)

var pki = testutil.NewPKI("../../../testutil/pki")

// Make sure X509Cert implements telegraf.Input
var _ telegraf.Input = &X509Cert{}

func TestGatherRemoteIntegration(t *testing.T) {
	t.Skip("Skipping network-dependent test due to race condition when test-all")

	tmpfile, err := ioutil.TempFile("", "example")
	require.NoError(t, err)

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(pki.ReadServerCert()))
	require.NoError(t, err)

	tests := []struct {
		name    string
		server  string
		timeout time.Duration
		close   bool
		unset   bool
		noshake bool
		error   bool
	}{
		{name: "wrong port", server: ":99999", error: true},
		{name: "no server", timeout: 5},
		{name: "successful https", server: "https://example.org:443", timeout: 5},
		{name: "successful file", server: "file://" + filepath.ToSlash(tmpfile.Name()), timeout: 5},
		{name: "unsupported scheme", server: "foo://", timeout: 5, error: true},
		{name: "no certificate", timeout: 5, unset: true, error: true},
		{name: "closed connection", close: true, error: true},
		{name: "no handshake", timeout: 5, noshake: true, error: true},
	}

	pair, err := tls.X509KeyPair([]byte(pki.ReadServerCert()), []byte(pki.ReadServerKey()))
	require.NoError(t, err)

	cfg := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{pair},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.unset {
				cfg.Certificates = nil
				cfg.GetCertificate = func(i *tls.ClientHelloInfo) (*tls.Certificate, error) {
					return nil, nil
				}
			}

			ln, err := tls.Listen("tcp", ":0", cfg)
			require.NoError(t, err)
			defer ln.Close()

			go func() {
				sconn, err := ln.Accept()
				require.NoError(t, err)
				if test.close {
					sconn.Close()
				}

				serverConfig := cfg.Clone()

				srv := tls.Server(sconn, serverConfig)
				if test.noshake {
					srv.Close()
				}
				require.NoError(t, srv.Handshake())
			}()

			if test.server == "" {
				test.server = "tcp://" + ln.Addr().String()
			}

			sc := X509Cert{
				Sources: []string{test.server},
				Timeout: config.Duration(test.timeout),
			}
			require.NoError(t, sc.Init())

			sc.InsecureSkipVerify = true
			testErr := false

			acc := testutil.Accumulator{}
			err = sc.Gather(&acc)
			if len(acc.Errors) > 0 {
				testErr = true
			}

			if testErr != test.error {
				t.Errorf("%s", err)
			}
		})
	}
}

func TestGatherLocal(t *testing.T) {
	wrongCert := fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----\n", base64.StdEncoding.EncodeToString([]byte("test")))

	tests := []struct {
		name    string
		mode    os.FileMode
		content string
		error   bool
	}{
		{name: "permission denied", mode: 0001, error: true},
		{name: "not a certificate", mode: 0640, content: "test", error: true},
		{name: "wrong certificate", mode: 0640, content: wrongCert, error: true},
		{name: "correct certificate", mode: 0640, content: pki.ReadServerCert()},
		{name: "correct client certificate", mode: 0640, content: pki.ReadClientCert()},
		{name: "correct certificate and extra trailing space", mode: 0640, content: pki.ReadServerCert() + " "},
		{name: "correct certificate and extra leading space", mode: 0640, content: " " + pki.ReadServerCert()},
		{name: "correct multiple certificates", mode: 0640, content: pki.ReadServerCert() + pki.ReadCACert()},
		{name: "correct multiple certificates and key", mode: 0640, content: pki.ReadServerCert() + pki.ReadCACert() + pki.ReadServerKey()},
		{name: "correct certificate and wrong certificate", mode: 0640, content: pki.ReadServerCert() + "\n" + wrongCert, error: true},
		{name: "correct certificate and not a certificate", mode: 0640, content: pki.ReadServerCert() + "\ntest", error: true},
		{name: "correct multiple certificates and extra trailing space", mode: 0640, content: pki.ReadServerCert() + pki.ReadServerCert() + " "},
		{name: "correct multiple certificates and extra leading space", mode: 0640, content: " " + pki.ReadServerCert() + pki.ReadServerCert()},
		{name: "correct multiple certificates and extra middle space", mode: 0640, content: pki.ReadServerCert() + " " + pki.ReadServerCert()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, err := ioutil.TempFile("", "x509_cert")
			require.NoError(t, err)

			_, err = f.Write([]byte(test.content))
			require.NoError(t, err)

			if runtime.GOOS != "windows" {
				require.NoError(t, f.Chmod(test.mode))
			}

			require.NoError(t, f.Close())

			defer os.Remove(f.Name())

			sc := X509Cert{
				Sources: []string{f.Name()},
			}
			require.NoError(t, sc.Init())

			acc := testutil.Accumulator{}
			err = sc.Gather(&acc)

			if (len(acc.Errors) > 0) != test.error {
				t.Errorf("%s", err)
			}
		})
	}
}

func TestTags(t *testing.T) {
	cert := fmt.Sprintf("%s\n%s", pki.ReadServerCert(), pki.ReadCACert())

	f, err := ioutil.TempFile("", "x509_cert")
	require.NoError(t, err)

	_, err = f.Write([]byte(cert))
	require.NoError(t, err)

	require.NoError(t, f.Close())

	defer os.Remove(f.Name())

	sc := X509Cert{
		Sources: []string{f.Name()},
	}
	require.NoError(t, sc.Init())

	acc := testutil.Accumulator{}
	require.NoError(t, sc.Gather(&acc))

	assert.True(t, acc.HasMeasurement("x509_cert"))

	assert.True(t, acc.HasTag("x509_cert", "common_name"))
	assert.Equal(t, "server.localdomain", acc.TagValue("x509_cert", "common_name"))

	assert.True(t, acc.HasTag("x509_cert", "signature_algorithm"))
	assert.Equal(t, "SHA256-RSA", acc.TagValue("x509_cert", "signature_algorithm"))

	assert.True(t, acc.HasTag("x509_cert", "public_key_algorithm"))
	assert.Equal(t, "RSA", acc.TagValue("x509_cert", "public_key_algorithm"))

	assert.True(t, acc.HasTag("x509_cert", "issuer_common_name"))
	assert.Equal(t, "Telegraf Test CA", acc.TagValue("x509_cert", "issuer_common_name"))

	assert.True(t, acc.HasTag("x509_cert", "san"))
	assert.Equal(t, "localhost,127.0.0.1", acc.TagValue("x509_cert", "san"))

	assert.True(t, acc.HasTag("x509_cert", "serial_number"))
	serialNumber := new(big.Int)
	_, validSerialNumber := serialNumber.SetString(acc.TagValue("x509_cert", "serial_number"), 16)
	if !validSerialNumber {
		t.Errorf("Expected a valid Hex serial number but got %s", acc.TagValue("x509_cert", "serial_number"))
	}
	assert.Equal(t, big.NewInt(1), serialNumber)
}

func TestGatherChain(t *testing.T) {
	cert := fmt.Sprintf("%s\n%s", pki.ReadServerCert(), pki.ReadCACert())

	tests := []struct {
		name    string
		content string
		error   bool
	}{
		{name: "chain certificate", content: cert},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f, err := ioutil.TempFile("", "x509_cert")
			require.NoError(t, err)

			_, err = f.Write([]byte(test.content))
			require.NoError(t, err)

			require.NoError(t, f.Close())

			defer os.Remove(f.Name())

			sc := X509Cert{
				Sources: []string{f.Name()},
			}
			require.NoError(t, sc.Init())

			acc := testutil.Accumulator{}
			err = sc.Gather(&acc)
			if (err != nil) != test.error {
				t.Errorf("%s", err)
			}
		})
	}
}

func TestStrings(t *testing.T) {
	sc := X509Cert{}
	require.NoError(t, sc.Init())

	tests := []struct {
		name     string
		method   string
		returned string
		expected string
	}{
		{name: "description", method: "Description", returned: sc.Description(), expected: description},
		{name: "sample config", method: "SampleConfig", returned: sc.SampleConfig(), expected: sampleConfig},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.returned != test.expected {
				t.Errorf("Expected method %s to return '%s', found '%s'.", test.method, test.expected, test.returned)
			}
		})
	}
}

func TestGatherCertIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	m := &X509Cert{
		Sources: []string{"https://www.influxdata.com:443"},
	}
	require.NoError(t, m.Init())

	var acc testutil.Accumulator
	require.NoError(t, m.Gather(&acc))

	assert.True(t, acc.HasMeasurement("x509_cert"))
}

func TestGatherCertMustNotTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	duration := time.Duration(15) * time.Second
	m := &X509Cert{
		Sources: []string{"https://www.influxdata.com:443"},
		Timeout: config.Duration(duration),
	}
	require.NoError(t, m.Init())

	var acc testutil.Accumulator
	require.NoError(t, m.Gather(&acc))
	require.Empty(t, acc.Errors)
	assert.True(t, acc.HasMeasurement("x509_cert"))
}

func TestServerName(t *testing.T) {
	tests := []struct {
		name     string
		fromTLS  string
		fromCfg  string
		url      string
		expected string
		err      bool
	}{
		{name: "in cfg", fromCfg: "example.com", url: "https://other.example.com", expected: "example.com"},
		{name: "in tls", fromTLS: "example.com", url: "https://other.example.com", expected: "example.com"},
		{name: "from URL", url: "https://other.example.com", expected: "other.example.com"},
		{name: "errors", fromCfg: "otherex.com", fromTLS: "example.com", url: "https://other.example.com", err: true},
	}

	for _, elt := range tests {
		test := elt
		t.Run(test.name, func(t *testing.T) {
			sc := &X509Cert{
				ServerName:   test.fromCfg,
				ClientConfig: _tls.ClientConfig{ServerName: test.fromTLS},
			}
			require.NoError(t, sc.Init())
			u, err := url.Parse(test.url)
			require.NoError(t, err)
			actual, err := sc.serverName(u)
			if test.err {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, actual)
		})
	}
}
