package x509_cert

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

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

	tests := []struct {
		server  string
		timeout time.Duration
		close   bool
		unset   bool
		noshake bool
		error   bool
	}{
		{server: ":99999", timeout: 0, close: false, unset: false, error: true},
		{server: "", timeout: 5, close: false, unset: false, error: false},
		{server: "", timeout: 5, close: false, unset: true, error: true},
		{server: "", timeout: 0, close: true, unset: false, error: true},
		{server: "", timeout: 5, close: false, unset: false, noshake: true, error: true},
	}

	pair, err := tls.X509KeyPair([]byte(pki.ReadServerCert()), []byte(pki.ReadServerKey()))
	if err != nil {
		t.Fatal(err)
	}

	config := &tls.Config{
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{pair},
	}

	for i, test := range tests {
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

		sc.InsecureSkipVerify = true
		testErr := false

		acc := testutil.Accumulator{}
		err = sc.Gather(&acc)
		if err != nil {
			testErr = true
		}

		if testErr != test.error {
			t.Errorf("Test [%d]: %s.", i, err)
		}
	}
}

func TestGatherLocal(t *testing.T) {
	wrongCert := fmt.Sprintf("-----BEGIN CERTIFICATE-----\n%s\n-----END CERTIFICATE-----\n", base64.StdEncoding.EncodeToString([]byte("test")))

	tests := []struct {
		mode    os.FileMode
		content string
		error   bool
	}{
		{mode: 0001, content: "", error: true},
		{mode: 0640, content: "test", error: true},
		{mode: 0640, content: wrongCert, error: true},
		{mode: 0640, content: pki.ReadServerCert(), error: false},
	}

	for i, test := range tests {
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

		error := false

		acc := testutil.Accumulator{}
		err = sc.Gather(&acc)
		if err != nil {
			error = true
		}

		if error != test.error {
			t.Errorf("Test [%d]: %s.", i, err)
		}
	}
}

func TestStrings(t *testing.T) {
	sc := X509Cert{}

	tests := []struct {
		method   string
		returned string
		expected string
	}{
		{method: "Description", returned: sc.Description(), expected: description},
		{method: "SampleConfig", returned: sc.SampleConfig(), expected: sampleConfig},
	}

	for i, test := range tests {
		if test.returned != test.expected {
			t.Errorf("Test [%d]: Expected method %s to return '%s', found '%s'.", i, test.method, test.expected, test.returned)
		}
	}
}
