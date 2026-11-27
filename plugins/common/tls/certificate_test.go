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

func TestCompileCertificateFilterErrors(t *testing.T) {
	tests := []struct {
		name     string
		allowed  []string
		expected string
	}{
		{
			name:     "missing type separator",
			allowed:  []string{"client-1"},
			expected: `invalid filter expression "client-1"`,
		},
		{
			name:     "unknown type",
			allowed:  []string{"foo:bar"},
			expected: `invalid type "foo"`,
		},
		{
			name:     "invalid glob",
			allowed:  []string{"cn:["},
			expected: `compiling "cn" filter failed`,
		},
		{
			name:     "invalid CIDR",
			allowed:  []string{"ip:10.0.0.1/a"},
			expected: `parsing CIDR`,
		},
		{
			name:     "invalid CIDR after valid one",
			allowed:  []string{"ip:10.0.0.0/8", "ip:foo"},
			expected: `parsing CIDR`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compileCertificateFilter(tt.allowed)
			require.ErrorContains(t, err, tt.expected)
		})
	}
}

func TestCertificateFilterMatch(t *testing.T) {
	u, err := url.Parse("spiffe://example.org/ns/prod")
	require.NoError(t, err)
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(255),
		Subject: pkix.Name{
			CommonName:         "client-1",
			Organization:       []string{"Org1", "Org2"},
			OrganizationalUnit: []string{"Unit1", "Unit2"},
		},
		Issuer:         pkix.Name{CommonName: "My CA"},
		DNSNames:       []string{"a.example.org", "b.example.org"},
		EmailAddresses: []string{"user@example.org"},
		IPAddresses:    []net.IP{net.ParseIP("10.0.0.1"), net.ParseIP("fe80::1")},
		URIs:           []*url.URL{u},
	}

	tests := []struct {
		name     string
		cert     *x509.Certificate
		allowed  []string
		expected bool
	}{
		{
			name:     "no filters",
			expected: false,
		},
		{
			name:     "cn exact",
			allowed:  []string{"cn:client-1"},
			expected: true,
		},
		{
			name:     "cn glob",
			allowed:  []string{"cn:client-*"},
			expected: true,
		},
		{
			name:     "cn mismatch",
			allowed:  []string{"cn:client-2"},
			expected: false,
		},
		{
			name:     "org second value",
			allowed:  []string{"org:Org2"},
			expected: true,
		},
		{
			name:     "org mismatch",
			allowed:  []string{"org:Org3"},
			expected: false,
		},
		{
			name:     "unit second value",
			allowed:  []string{"unit:Unit2"},
			expected: true,
		},
		{
			name:     "unit mismatch",
			allowed:  []string{"unit:Unit3"},
			expected: false,
		},
		{
			name:     "issuer",
			allowed:  []string{"issuer:My CA"},
			expected: true,
		},
		{
			name:     "issuer does not match subject",
			allowed:  []string{"issuer:client-1"},
			expected: false,
		},
		{
			name:     "serial number hexadecimal",
			allowed:  []string{"sn:ff"},
			expected: true,
		},
		{
			name:     "serial number decimal",
			allowed:  []string{"sn:255"},
			expected: false,
		},
		{
			name:     "dns glob",
			allowed:  []string{"dns:*.example.org"},
			expected: true,
		},
		{
			name:     "dns mismatch",
			allowed:  []string{"dns:c.example.org"},
			expected: false,
		},
		{
			name:     "email",
			allowed:  []string{"email:*@example.org"},
			expected: true,
		},
		{
			name:     "email mismatch",
			allowed:  []string{"email:user@example.com"},
			expected: false,
		},
		{
			name:     "single ip v4",
			allowed:  []string{"ip:10.0.0.1"},
			expected: true,
		},
		{
			name:     "ip v4 range",
			allowed:  []string{"ip:10.0.0.0/8"},
			expected: true,
		},
		{
			name:     "ip v6 range",
			allowed:  []string{"ip:fe80::/10"},
			expected: true,
		},
		{
			name:     "ip mismatch",
			allowed:  []string{"ip:192.0.2.0/24"},
			expected: false,
		},
		{
			name:     "ip second range",
			allowed:  []string{"ip:192.0.2.0/24", "ip:10.0.0.0/24"},
			expected: true,
		},
		{
			name:     "uri",
			allowed:  []string{"uri:spiffe://example.org/ns/*"},
			expected: true,
		},
		{
			name:     "uri mismatch",
			allowed:  []string{"uri:spiffe://example.org/ns/dev"},
			expected: false,
		},
		{
			name:     "multiple expressions of same type",
			allowed:  []string{"cn:foo", "cn:client-1"},
			expected: true,
		},
		{
			name:     "any type matching is sufficient",
			allowed:  []string{"cn:foo", "dns:b.example.org"},
			expected: true,
		},
		{
			name:     "no type matching",
			allowed:  []string{"cn:foo", "dns:c.example.org", "ip:192.0.2.0/24"},
			expected: false,
		},
		{
			name:     "certificate without attributes",
			cert:     &x509.Certificate{},
			allowed:  []string{"org:*", "unit:*", "dns:*", "email:*", "ip:0.0.0.0/0", "uri:*"},
			expected: false,
		},
		{
			name: "range does not match other ip",
			cert: &x509.Certificate{
				SerialNumber: big.NewInt(1),
				IPAddresses:  []net.IP{net.ParseIP("192.0.2.5")},
			},
			allowed:  []string{"ip:10.0.0.0/8"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.cert
			if c == nil {
				c = cert
			}
			f, err := compileCertificateFilter(tt.allowed)
			require.NoError(t, err)
			require.Equal(t, tt.expected, f.match(c))
		})
	}
}
