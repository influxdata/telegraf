package tls

import (
	"crypto/x509"
	"fmt"
	"strings"
)

// CertificateTags extracts the requested attributes from the given
// certificate and returns them as tag-like key-value pairs. Attributes without
// a value are omitted. Multiple values of an attribute are joined by a comma.
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
//
// An error is returned for unknown attribute names, so calling the function
// with an empty certificate can be used to validate a list of names.
func CertificateTags(cert *x509.Certificate, names []string) (map[string]string, error) {
	tags := make(map[string]string, len(names))
	for _, name := range names {
		var value string
		switch name {
		case "cn":
			value = cert.Subject.CommonName
		case "org":
			value = strings.Join(cert.Subject.Organization, ",")
		case "unit":
			value = strings.Join(cert.Subject.OrganizationalUnit, ",")
		case "issuer":
			value = cert.Issuer.CommonName
		case "sn":
			if cert.SerialNumber != nil {
				value = cert.SerialNumber.Text(16)
			}
		case "dns":
			value = strings.Join(cert.DNSNames, ",")
		case "email":
			value = strings.Join(cert.EmailAddresses, ",")
		case "ip":
			ips := make([]string, 0, len(cert.IPAddresses))
			for _, ip := range cert.IPAddresses {
				ips = append(ips, ip.String())
			}
			value = strings.Join(ips, ",")
		case "uri":
			uris := make([]string, 0, len(cert.URIs))
			for _, u := range cert.URIs {
				uris = append(uris, u.String())
			}
			value = strings.Join(uris, ",")
		default:
			return nil, fmt.Errorf("invalid certificate attribute %q", name)
		}

		if value != "" {
			tags[name] = value
		}
	}
	return tags, nil
}
