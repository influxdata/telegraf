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

	tests := []struct {
		name     string
		cert     *x509.Certificate
		names    []string
		prefix   string
		expected map[string]string
	}{
		{
			name:  "all attributes",
			cert:  cert,
			names: all,
			expected: map[string]string{
				"cn":     "client-1",
				"org":    "Org",
				"unit":   "Unit1,Unit2",
				"issuer": "My CA",
				"sn":     "ff",
				"dns":    "a.example.org,b.example.org",
				"email":  "user@example.org",
				"ip":     "10.0.0.1,fe80::1",
				"uri":    "spiffe://example.org/ns/prod",
			},
		},
		{
			name:     "subset of attributes",
			cert:     cert,
			names:    []string{"cn", "dns"},
			expected: map[string]string{"cn": "client-1", "dns": "a.example.org,b.example.org"},
		},
		{
			name:     "with prefix",
			cert:     cert,
			names:    []string{"cn", "sn"},
			prefix:   "tls_client_",
			expected: map[string]string{"tls_client_cn": "client-1", "tls_client_sn": "ff"},
		},
		{
			name:     "no names",
			cert:     cert,
			expected: map[string]string{},
		},
		{
			name:     "unknown names are ignored",
			cert:     cert,
			names:    []string{"cn", "foo"},
			expected: map[string]string{"cn": "client-1"},
		},
		{
			name:     "empty certificate",
			cert:     &x509.Certificate{},
			names:    all,
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, CertificateTags(tt.cert, tt.names, tt.prefix))
		})
	}
}

func TestValidateTagNames(t *testing.T) {
	tests := []struct {
		name     string
		names    []string
		expected string
	}{
		{
			name:  "all valid",
			names: []string{"cn", "org", "unit", "issuer", "sn", "dns", "email", "ip", "uri"},
		},
		{
			name: "empty",
		},
		{
			name:     "invalid name",
			names:    []string{"cn", "foo"},
			expected: `invalid certificate attribute "foo"`,
		},
		{
			name:     "old style name",
			names:    []string{"common_name"},
			expected: `invalid certificate attribute "common_name"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTagNames(tt.names)
			if tt.expected != "" {
				require.ErrorContains(t, err, tt.expected)
				return
			}
			require.NoError(t, err)
		})
	}
}
