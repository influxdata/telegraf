package opentelemetry

import (
	"context"
	"crypto/x509"
	"fmt"
	"slices"
	"strings"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// Supported values of the `tls_client_cert_tags` option.
var clientCertTagNames = []string{
	"common_name",
	"organization",
	"organizational_unit",
	"issuer_common_name",
	"serial_number",
	"san_dns",
	"san_email",
	"san_ip",
	"san_uri",
}

func checkClientCertTags(names []string) error {
	for _, n := range names {
		if !slices.Contains(clientCertTagNames, n) {
			return fmt.Errorf("invalid client certificate tag %q, valid values are: %s", n, strings.Join(clientCertTagNames, ", "))
		}
	}
	return nil
}

// clientCertTags returns the requested tags derived from the certificate the
// client presented during the (mutual) TLS handshake. It returns nil if there
// is no such certificate, e.g. for plain-text or server-only TLS connections.
func clientCertTags(ctx context.Context, names []string) map[string]string {
	if len(names) == 0 {
		return nil
	}
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil
	}
	info, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok || len(info.State.PeerCertificates) == 0 {
		return nil
	}
	cert := info.State.PeerCertificates[0]

	tags := make(map[string]string, len(names))
	for _, n := range names {
		var v string
		switch n {
		case "common_name":
			v = cert.Subject.CommonName
		case "organization":
			v = strings.Join(cert.Subject.Organization, ",")
		case "organizational_unit":
			v = strings.Join(cert.Subject.OrganizationalUnit, ",")
		case "issuer_common_name":
			v = cert.Issuer.CommonName
		case "serial_number":
			v = cert.SerialNumber.Text(16)
		case "san_dns":
			v = strings.Join(cert.DNSNames, ",")
		case "san_email":
			v = strings.Join(cert.EmailAddresses, ",")
		case "san_ip":
			v = joinIPs(cert)
		case "san_uri":
			uris := make([]string, 0, len(cert.URIs))
			for _, u := range cert.URIs {
				uris = append(uris, u.String())
			}
			v = strings.Join(uris, ",")
		}
		// Do not emit empty tags
		if v != "" {
			tags["tls_client_"+n] = v
		}
	}
	return tags
}

func joinIPs(cert *x509.Certificate) string {
	ips := make([]string, 0, len(cert.IPAddresses))
	for _, ip := range cert.IPAddresses {
		ips = append(ips, ip.String())
	}
	return strings.Join(ips, ",")
}

// addClientCertTags adds the certificate tags to the given tag-map. Tags
// derived from the certificate take precedence over telemetry attributes with
// the same name as those are controlled by the sender.
func addClientCertTags(ctx context.Context, names []string, tags map[string]string) map[string]string {
	extra := clientCertTags(ctx, names)
	if len(extra) == 0 {
		return tags
	}
	if tags == nil {
		tags = make(map[string]string, len(extra))
	}
	for k, v := range extra {
		tags[k] = v
	}
	return tags
}
