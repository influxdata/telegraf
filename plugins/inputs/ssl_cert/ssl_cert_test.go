package ssl_cert

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
)

const testCert = `-----BEGIN CERTIFICATE-----
MIID7DCCAtSgAwIBAgIBATANBgkqhkiG9w0BAQsFADB1MQswCQYDVQQGEwJVSzEX
MBUGA1UECBMOVW5pdGVkIEtpbmdkb20xDzANBgNVBAcTBkxvbmRvbjEWMBQGA1UE
ChMNVGVsZWdyYWYgVGVzdDEQMA4GA1UECxMHVGVzdGluZzESMBAGA1UEAxMJbG9j
YWxob3N0MB4XDTE4MDIwNjA2MTIwMFoXDTE5MDIwNjA2MTIwMFowdTELMAkGA1UE
BhMCVUsxFzAVBgNVBAgTDlVuaXRlZCBLaW5nZG9tMQ8wDQYDVQQHEwZMb25kb24x
FjAUBgNVBAoTDVRlbGVncmFmIFRlc3QxEDAOBgNVBAsTB1Rlc3RpbmcxEjAQBgNV
BAMTCWxvY2FsaG9zdDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKev
IBsUu3NdYod/jxPxpJug4p1M/qiGkFTffcrjbxzlZUIED5bAZNUutqVYQYT7/Chy
pt7U7B6coAbUoIbJLpQ9ktDcyF22LpV7E6TB2SFIemnNYPM0U9vcW2EEaUvFfCTT
RLiwd+S4/twfgdgP8RuCnXuqkRQXCJVwD1GnazuNHqlyYH4IznXmY/Ia6dlaTSKs
zuJp4UwYSAuO+AmoefmOUAJ0Q2l4khBZlXv5wHqROuX+3VdIK7JtoP0ydcPkrTt9
zI/qt4AFBJvIdBZS9QVAwoBNnNAElHX5eRRWhuLgftHVQM+3JNpsXYp4OlcoAs/A
QowOSsKXJu4i8pIn7lkCAwEAAaOBhjCBgzAMBgNVHRMBAf8EAjAAMB0GA1UdDgQW
BBRfip83ocgzVxMCnsFHASt4CUY5qjALBgNVHQ8EBAMCBeAwFAYDVR0RBA0wC4IJ
bG9jYWxob3N0MBEGCWCGSAGG+EIBAQQEAwIGQDAeBglghkgBhvhCAQ0EERYPeGNh
IGNlcnRpZmljYXRlMA0GCSqGSIb3DQEBCwUAA4IBAQBOuCcjgiDL9OhSuLN3pewB
tE0garxFgPBGwkr8g+XpRktiensdhoxsQFwuqsIPv56GmqlqiqZb98HxnNFkQ0im
XgmHvSxHLiSKbD6DOgpKQ1I0FB7sHKb8qaJbxwPJDXMyXseja7PEzh5EVtQMimhQ
z46Fz+pZ4uUS2h2TgqiAoJpKZnb8mceGQY9qzGztWv67RFsWVK7R+JP0W2cZ+Tk+
2NoeMhBdqV0fA13cFvvWSLZ2w3eKLldJMIni2hk/G9nHDWIAbGb83qdhGfR7PMkJ
jD/IOPXIcYxKB62aDCEWDcCGdLZCRgBP2VDzQW2J6MIFm6qv5H8llkNs3Qh1/VFj
-----END CERTIFICATE-----`

const testKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAp68gGxS7c11ih3+PE/Gkm6DinUz+qIaQVN99yuNvHOVlQgQP
lsBk1S62pVhBhPv8KHKm3tTsHpygBtSghskulD2S0NzIXbYulXsTpMHZIUh6ac1g
8zRT29xbYQRpS8V8JNNEuLB35Lj+3B+B2A/xG4Kde6qRFBcIlXAPUadrO40eqXJg
fgjOdeZj8hrp2VpNIqzO4mnhTBhIC474Cah5+Y5QAnRDaXiSEFmVe/nAepE65f7d
V0grsm2g/TJ1w+StO33Mj+q3gAUEm8h0FlL1BUDCgE2c0ASUdfl5FFaG4uB+0dVA
z7ck2mxding6VygCz8BCjA5Kwpcm7iLykifuWQIDAQABAoIBADZeOLmvGiwIjkbC
nCBqS+XN30wDR9pabvel0wJyhXdIBXHHIUrOrKLWV4/6spusnBB9RA+h18EBJX2x
eS7akgisgirIOwrvY+FBm5fi5kS9XDtrxNB2Ge6CXvpw1Lclm9/QxEphpS36sV+r
s4zbdmBmFCuhnRJ3eWgCgmUGNGWFEG2hmYAT5cy1ViPRY5pBNMLz+zvWhDw1Zjwl
C7Oe/zqqc7R1vGmoUJ8KPHqPAkXF9Ouj61YT+WXrF1iRRhn4oJ7x9O4L1vkbYOxq
0PWh5gGGHcqe3SdvF6RI4AlhLBK1PNPeFU78k4VXBZoeBNg+Q7lQYYr1A56sE+kY
BbqGAdECgYEA2nq4GlcbzBSeH3NzAkxy+lPYgdswrGih7mJeCwmJby+IdkqK5DL9
VyC95Bx2aut0hiv8fxVfAM+utWn5tyfBt1BVhtyMj5Fam8Deh/JjgmJi4GeaWcMg
lo4iTUFTOBQ5ibGzUeE8pOnViN0HkUwpzWE2cvWh1546UgQkQJCAHU0CgYEAxHs6
zv4tCz1g9Hj5nlRINdTbMJwr2Es1dNVzYeDlvtU7IfT/82G/Z4Ub35QF0vcj/5uB
sBKrF6i023WgDpQ/pqNoIE4GpEzyNov07e4YoYU9ejBamBdI48gmJswegYRjCeHf
Gl/cMBbSYgTA+GKsn2I5cWIVAEu50f8XUqSXPz0CgYEAnc2VvDC+uyEJNN5Ga5qc
UYLOFr0i4uSQUYZrNr2krtI+VnJw73KE2bGkdma4gXGfsGmE7qWZARUAs7ffzhLB
MI6tt8MFI41xTJ56HOdOSJaXpE4whjUSDKyMyhAs84xoIrRfOPzeuJ7MxRYgqSnB
574XfeE9DGgU57hmFtxILOECgYEAm+f4kzVHQsryeyrfT84q+mQrhVf2xotvIIUb
KEiPpSyH3nsM+e/PNHJ/2poXQP6QVwvrDW7SylQ5JocgeVETbMPvJOslBAx2iefm
c0Hh05DpZmKmEFcxpGU2OMTxU+5btATBxqjYDGSfjd2dzbpmpZYIZLriVTjBeyuC
MzadOTUCgYEAk75Ioeb7zURwvgXbP25MUlm1dTE+vT1Q63QSRHAgMqVZ6seexv09
JoiUjTAZegW1RkST3tB1an9zO0EcPvo/1yU7wKaMuNwmauPlltBdpaTwojeUBiFr
AoQUXWIiFBoFfpNVcvUgHyGGc1hv7TX8Eh8KGo2+VuPzxnFuRrbbYCs=
-----END RSA PRIVATE KEY-----`

// Make sure SSLCert implements telegraf.Input
var _ telegraf.Input = &SSLCert{}

func TestGatherRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network-dependent test in short mode.")
	}

	tests := []struct {
		server  string
		timeout time.Duration
		close   bool
		unset   bool
		error   bool
	}{
		{server: ":99999", timeout: 0, close: false, unset: false, error: true},
		{server: "", timeout: 5, close: false, unset: false, error: false},
		{server: "", timeout: 5, close: false, unset: true, error: true},
		{server: "", timeout: 0, close: true, unset: false, error: true},
	}

	for i, test := range tests {
		pair, err := tls.X509KeyPair([]byte(testCert), []byte(testKey))
		if err != nil {
			t.Fatal(err)
		}

		config := &tls.Config{
			Certificates: []tls.Certificate{pair},
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

			serverConfig := config.Clone()

			srv := tls.Server(sconn, serverConfig)
			if err := srv.Handshake(); err != nil {
				return
			}
		}()

		if test.server == "" {
			test.server = ln.Addr().String()
		}

		sc := SSLCert{
			Servers:    []string{test.server},
			Timeout:    test.timeout,
			CloseConn:  test.close,
			UnsetCerts: test.unset,
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
		{mode: 0640, content: testCert, error: false},
	}

	for i, test := range tests {
		f, err := ioutil.TempFile("", "ssl_cert")
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

		sc := SSLCert{
			Files: []string{f.Name()},
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
	sc := SSLCert{}

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
