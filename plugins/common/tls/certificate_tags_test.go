package tls

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCertificateTags(t *testing.T) {
	u, err := url.Parse("spiffe://example.org/ns/prod")
	require.NoError(t, err)
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(255),
		Subject: pkix.Name{
			CommonName:         "client-1",
			Organization:       []string{"Org"},
			OrganizationalUnit: []string{"Unit1", "Unit2"},
		},
		Issuer:         pkix.Name{CommonName: "My CA"},
		DNSNames:       []string{"a.example.org", "b.example.org"},
		EmailAddresses: []string{"user@example.org"},
		IPAddresses:    []net.IP{net.ParseIP("10.0.0.1"), net.ParseIP("fe80::1")},
		URIs:           []*url.URL{u},
	}

	all := []string{"cn", "org", "unit", "issuer", "sn", "dns", "email", "ip", "uri"}
	expected := map[string]string{
		"cn":     "client-1",
		"org":    "Org",
		"unit":   "Unit1,Unit2",
		"issuer": "My CA",
		"sn":     "ff",
		"dns":    "a.example.org,b.example.org",
		"email":  "user@example.org",
		"ip":     "10.0.0.1,fe80::1",
		"uri":    "spiffe://example.org/ns/prod",
	}
	actual, err := CertificateTags(cert, all)
	require.NoError(t, err)
	require.Equal(t, expected, actual)

	// Only the requested attributes are returned
	actual, err = CertificateTags(cert, []string{"cn", "dns"})
	require.NoError(t, err)
	require.Equal(t, map[string]string{"cn": "client-1", "dns": "a.example.org,b.example.org"}, actual)

	// No names
	actual, err = CertificateTags(cert, nil)
	require.NoError(t, err)
	require.Empty(t, actual)
}

func TestCertificateTagsEmptyCertificate(t *testing.T) {
	all := []string{"cn", "org", "unit", "issuer", "sn", "dns", "email", "ip", "uri"}
	actual, err := CertificateTags(&x509.Certificate{}, all)
	require.NoError(t, err)
	require.Empty(t, actual)
}

func TestCertificateTagsInvalid(t *testing.T) {
	_, err := CertificateTags(&x509.Certificate{}, []string{"cn", "foo"})
	require.ErrorContains(t, err, `invalid certificate attribute "foo"`)
}
