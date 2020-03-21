package x509_cert

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/testutil"
)

var pki = testutil.NewPKI("../../../testutil/pki")

// Make sure X509Cert implements telegraf.Input
var _ telegraf.Input = &X509Cert{}

func TestGatherRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}

	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(pki.ReadServerCert())); err != nil {
		t.Fatal(err)
	}

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
		{name: "successful file", server: "file://" + tmpfile.Name(), timeout: 5},
		{name: "unsupported scheme", server: "foo://", timeout: 5, error: true},
		{name: "no certificate", timeout: 5, unset: true, error: true},
		{name: "closed connection", close: true, error: true},
		{name: "no handshake", timeout: 5, noshake: true, error: true},
	}

	pair, err := tls.X509KeyPair([]byte(pki.ReadServerCert()), []byte(pki.ReadServerKey()))
	if err != nil {
		t.Fatal(err)
	}

	config := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{pair},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.unset {
				config.Certificates = nil
				config.GetCertificate = func(i *tls.ClientHelloInfo) (*tls.Certificate, error) {
					return nil, nil
				}
			}

			ln, err := tls.Listen("tcp", ":0", config)
			if err != nil {
				t.Fatal(err)
			}
			defer ln.Close()

			go func() {
				sconn, err := ln.Accept()
				if err != nil {
					return
				}
				if test.close {
					sconn.Close()
				}

				serverConfig := config.Clone()

				srv := tls.Server(sconn, serverConfig)
				if test.noshake {
					srv.Close()
				}
				if err := srv.Handshake(); err != nil {
					return
				}
			}()

			if test.server == "" {
				test.server = "tcp://" + ln.Addr().String()
			}

			sc := X509Cert{
				Sources: []string{test.server},
				Timeout: internal.Duration{Duration: test.timeout},
			}
			sc.Init()

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
			if err != nil {
				t.Fatal(err)
			}

			_, err = f.Write([]byte(test.content))
			if err != nil {
				t.Fatal(err)
			}

			err = f.Chmod(test.mode)
			if err != nil {
				t.Fatal(err)
			}

			err = f.Close()
			if err != nil {
				t.Fatal(err)
			}

			defer os.Remove(f.Name())

			sc := X509Cert{
				Sources: []string{f.Name()},
			}
			sc.Init()

			error := false

			acc := testutil.Accumulator{}
			err = sc.Gather(&acc)
			if len(acc.Errors) > 0 {
				error = true
			}

			if error != test.error {
				t.Errorf("%s", err)
			}
		})
	}
}

func TestTags(t *testing.T) {
	cert := fmt.Sprintf("%s\n%s", pki.ReadServerCert(), pki.ReadCACert())

	f, err := ioutil.TempFile("", "x509_cert")
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.Write([]byte(cert))
	if err != nil {
		t.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(f.Name())

	sc := X509Cert{
		Sources: []string{f.Name()},
	}
	sc.Init()

	acc := testutil.Accumulator{}
	err = sc.Gather(&acc)
	require.NoError(t, err)

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
			if err != nil {
				t.Fatal(err)
			}

			_, err = f.Write([]byte(test.content))
			if err != nil {
				t.Fatal(err)
			}

			err = f.Close()
			if err != nil {
				t.Fatal(err)
			}

			defer os.Remove(f.Name())

			sc := X509Cert{
				Sources: []string{f.Name()},
			}
			sc.Init()

			error := false

			acc := testutil.Accumulator{}
			err = sc.Gather(&acc)
			if err != nil {
				error = true
			}

			if error != test.error {
				t.Errorf("%s", err)
			}
		})
	}

}

func TestStrings(t *testing.T) {
	sc := X509Cert{}
	sc.Init()

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

func TestGatherCert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	m := &X509Cert{
		Sources: []string{"https://www.influxdata.com:443"},
	}
	m.Init()

	var acc testutil.Accumulator
	err := m.Gather(&acc)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("x509_cert"))
}
