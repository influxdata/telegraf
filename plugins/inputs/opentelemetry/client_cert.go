package opentelemetry

import (
	"context"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

// clientCertTags returns the requested tags, prefixed with "tls_client_",
// derived from the certificate the client presented during the (mutual) TLS
// handshake. It returns nil if there is no such certificate, e.g. for
// plain-text or server-only TLS connections. The names must be validated
// beforehand, see Init().
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

	certTags, err := tls.CertificateTags(info.State.PeerCertificates[0], names)
	if err != nil {
		return nil
	}
	tags := make(map[string]string, len(certTags))
	for k, v := range certTags {
		tags["tls_client_"+k] = v
	}
	return tags
}
