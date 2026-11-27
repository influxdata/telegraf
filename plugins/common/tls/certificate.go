package tls

import (
	"crypto/x509"
	"fmt"
	"maps"
	"net"
	"net/url"
	"slices"
	"strings"

	"github.com/influxdata/telegraf/filter"
)

type certificateFilter struct {
	clientCN       filter.Filter
	clientOrg      filter.Filter
	clientUnit     filter.Filter
	clientIssuer   filter.Filter
	clientSN       filter.Filter
	clientDNS      filter.Filter
	clientEmail    filter.Filter
	clientIPs      filter.Filter
	clientIPRanges []*net.IPNet
	clientURI      filter.Filter
}

func compileCertificateFilter(allowed []string) (*certificateFilter, error) {
	var c certificateFilter

	// Collect the filter expressions by type
	filterLists := make(map[string][]string)
	for _, f := range allowed {
		kind, expr, found := strings.Cut(f, ":")
		if !found {
			return nil, fmt.Errorf("invalid filter expression %q", f)
		}
		filterLists[kind] = append(filterLists[kind], expr)
	}

	// Check the filter types and compile the filters
	for k, l := range filterLists {
		switch k {
		case "cn":
			f, err := filter.Compile(l)
			if err != nil {
				return nil, fmt.Errorf("compiling %q filter failed: %w", k, err)
			}
			c.clientCN = f
		case "org":
			f, err := filter.Compile(l)
			if err != nil {
				return nil, fmt.Errorf("compiling %q filter failed: %w", k, err)
			}
			c.clientOrg = f
		case "unit":
			f, err := filter.Compile(l)
			if err != nil {
				return nil, fmt.Errorf("compiling %q filter failed: %w", k, err)
			}
			c.clientUnit = f
		case "issuer":
			f, err := filter.Compile(l)
			if err != nil {
				return nil, fmt.Errorf("compiling %q filter failed: %w", k, err)
			}
			c.clientIssuer = f
		case "sn":
			f, err := filter.Compile(l)
			if err != nil {
				return nil, fmt.Errorf("compiling %q filter failed: %w", k, err)
			}
			c.clientSN = f
		case "dns":
			f, err := filter.Compile(l)
			if err != nil {
				return nil, fmt.Errorf("compiling %q filter failed: %w", k, err)
			}
			c.clientDNS = f
		case "email":
			f, err := filter.Compile(l)
			if err != nil {
				return nil, fmt.Errorf("compiling %q filter failed: %w", k, err)
			}
			c.clientEmail = f
		case "ip":
			var individualIPs []string
			for _, expr := range l {
				_, iprange, err := net.ParseCIDR(expr)
				if err == nil {
					c.clientIPRanges = append(c.clientIPRanges, iprange)
					continue
				}
				if net.ParseIP(expr) != nil {
					individualIPs = append(individualIPs, expr)
					continue
				}
				return nil, fmt.Errorf("parsing CIDR %q filter failed: %w", expr, err)
			}
			f, err := filter.Compile(individualIPs)
			if err != nil {
				return nil, fmt.Errorf("compiling individual %q filter failed: %w", k, err)
			}
			c.clientIPs = f
		case "uri":
			f, err := filter.Compile(l)
			if err != nil {
				return nil, fmt.Errorf("compiling %q filter failed: %w", k, err)
			}
			c.clientURI = f
		default:
			return nil, fmt.Errorf("invalid type %q", k)
		}
	}
	return &c, nil
}

func (c *certificateFilter) match(cert *x509.Certificate) bool {
	if c.clientCN != nil && c.clientCN.Match(cert.Subject.CommonName) {
		return true
	}
	if c.clientOrg != nil && slices.ContainsFunc(cert.Subject.Organization, func(v string) bool { return c.clientOrg.Match(v) }) {
		return true
	}
	if c.clientUnit != nil && slices.ContainsFunc(cert.Subject.OrganizationalUnit, func(v string) bool { return c.clientUnit.Match(v) }) {
		return true
	}
	if c.clientIssuer != nil && c.clientIssuer.Match(cert.Issuer.CommonName) {
		return true
	}
	if c.clientSN != nil && c.clientSN.Match(cert.SerialNumber.Text(16)) {
		return true
	}
	if c.clientDNS != nil && slices.ContainsFunc(cert.DNSNames, func(v string) bool { return c.clientDNS.Match(v) }) {
		return true
	}
	if c.clientEmail != nil && slices.ContainsFunc(cert.EmailAddresses, func(v string) bool { return c.clientEmail.Match(v) }) {
		return true
	}
	if c.clientIPs != nil && slices.ContainsFunc(cert.IPAddresses, func(v net.IP) bool { return c.clientIPs.Match(v.String()) }) {
		return true
	}
	for _, iprange := range c.clientIPRanges {
		if slices.ContainsFunc(cert.IPAddresses, iprange.Contains) {
			return true
		}
	}
	if c.clientURI != nil && slices.ContainsFunc(cert.URIs, func(u *url.URL) bool { return c.clientURI.Match(u.String()) }) {
		return true
	}

	return false
}

// ValidateTagNames checks if the given names are valid certificate attributes
// to be used with CertificateTags.
func ValidateTagNames(names []string) error {
	for _, name := range names {
		switch name {
		case "cn", "org", "unit", "issuer", "sn", "dns", "email", "ip", "uri":
		default:
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
		switch name {
		case "cn":
			tags[prefix+name] = cert.Subject.CommonName
		case "org":
			tags[prefix+name] = strings.Join(cert.Subject.Organization, ",")
		case "unit":
			tags[prefix+name] = strings.Join(cert.Subject.OrganizationalUnit, ",")
		case "issuer":
			tags[prefix+name] = cert.Issuer.CommonName
		case "sn":
			if cert.SerialNumber != nil {
				tags[prefix+name] = cert.SerialNumber.Text(16)
			}
		case "dns":
			tags[prefix+name] = strings.Join(cert.DNSNames, ",")
		case "email":
			tags[prefix+name] = strings.Join(cert.EmailAddresses, ",")
		case "ip":
			ips := make([]string, 0, len(cert.IPAddresses))
			for _, ip := range cert.IPAddresses {
				ips = append(ips, ip.String())
			}
			tags[prefix+name] = strings.Join(ips, ",")
		case "uri":
			uris := make([]string, 0, len(cert.URIs))
			for _, u := range cert.URIs {
				uris = append(uris, u.String())
			}
			tags[prefix+name] = strings.Join(uris, ",")
		}
	}

	// Remove tags with empty value
	maps.DeleteFunc(tags, func(_, v string) bool { return v == "" })

	return tags
}
