package tls

import (
	"crypto/x509"
	"fmt"
	"strings"
)

// ValidateTagNames checks if the given names are valid certificate attributes
// to be used with CertificateTags.
func ValidateTagNames(names []string) error {
	for _, name := range names {
		if _, ok := certificateAttribute(&x509.Certificate{}, name); !ok {
			return fmt.Errorf("invalid certificate attribute %q", name)
		}
	}
	return nil
}

// CertificateTags extracts the requested attributes from the given
// certificate and returns them as tag-like key-value pairs where the key is the
// attribute name prefixed with the given prefix. Attributes without a value
// are omitted and multiple values of an attribute are joined by a comma. The
// names must be validated beforehand using ValidateTagNames, unknown names are
// ignored.
//
// Valid attribute names are:
//   - cn:     subject common name
//   - org:    subject organization
//   - unit:   subject organizational unit
//   - issuer: issuer common name
//   - sn:     serial number (hexadecimal)
//   - dns:    subject alternative names of type DNS
//   - email:  subject alternative names of type email
//   - ip:     subject alternative names of type IP address
//   - uri:    subject alternative names of type URI
func CertificateTags(cert *x509.Certificate, names []string, prefix string) map[string]string {
	tags := make(map[string]string, len(names))
	for _, name := range names {
		if value, ok := certificateAttribute(cert, name); ok && value != "" {
			tags[prefix+name] = value
		}
	}
	return tags
}

func certificateAttribute(cert *x509.Certificate, name string) (string, bool) {
	switch name {
	case "cn":
		return cert.Subject.CommonName, true
	case "org":
		return strings.Join(cert.Subject.Organization, ","), true
	case "unit":
		return strings.Join(cert.Subject.OrganizationalUnit, ","), true
	case "issuer":
		return cert.Issuer.CommonName, true
	case "sn":
		if cert.SerialNumber == nil {
			return "", true
		}
		return cert.SerialNumber.Text(16), true
	case "dns":
		return strings.Join(cert.DNSNames, ","), true
	case "email":
		return strings.Join(cert.EmailAddresses, ","), true
	case "ip":
		ips := make([]string, 0, len(cert.IPAddresses))
		for _, ip := range cert.IPAddresses {
			ips = append(ips, ip.String())
		}
		return strings.Join(ips, ","), true
	case "uri":
		uris := make([]string, 0, len(cert.URIs))
		for _, u := range cert.URIs {
			uris = append(uris, u.String())
		}
		return strings.Join(uris, ","), true
	}
	return "", false
}
