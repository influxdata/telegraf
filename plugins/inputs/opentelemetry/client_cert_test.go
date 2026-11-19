package opentelemetry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	"github.com/influxdata/telegraf/testutil"
)

func peerContext(cert *x509.Certificate) context.Context {
	info := credentials.TLSInfo{}
	if cert != nil {
		info.State = tls.ConnectionState{PeerCertificates: []*x509.Certificate{cert}}
	}
	return peer.NewContext(context.Background(), &peer.Peer{AuthInfo: info})
}

func TestClientCertTagsInvalid(t *testing.T) {
	plugin := &OpenTelemetry{ClientCertTags: []string{"common_name", "foo"}}
	require.ErrorContains(t, plugin.Init(), `invalid client certificate tag "foo"`)
}

func TestClientCertTags(t *testing.T) {
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
		IPAddresses:    []net.IP{net.ParseIP("10.0.0.1")},
		URIs:           []*url.URL{u},
	}

	all := append([]string{}, clientCertTagNames...)
	expected := map[string]string{
		"tls_client_common_name":         "client-1",
		"tls_client_organization":        "Org",
		"tls_client_organizational_unit": "Unit1,Unit2",
		"tls_client_issuer_common_name":  "My CA",
		"tls_client_serial_number":       "ff",
		"tls_client_san_dns":             "a.example.org,b.example.org",
		"tls_client_san_email":           "user@example.org",
		"tls_client_san_ip":              "10.0.0.1",
		"tls_client_san_uri":             "spiffe://example.org/ns/prod",
	}
	require.Equal(t, expected, clientCertTags(peerContext(cert), all))

	// Only the requested tags and spoofed attributes get overridden
	tags := map[string]string{"tls_client_common_name": "spoofed", "other": "x"}
	got := addClientCertTags(peerContext(cert), []string{"common_name"}, tags)
	require.Equal(t, map[string]string{"tls_client_common_name": "client-1", "other": "x"}, got)

	// Nothing without options, peer, TLS or client certificate
	require.Nil(t, clientCertTags(peerContext(cert), nil))
	require.Nil(t, clientCertTags(context.Background(), all))
	require.Nil(t, clientCertTags(peerContext(nil), all))
	require.Equal(t, map[string]string{"other": "x"}, addClientCertTags(peerContext(nil), all, map[string]string{"other": "x"}))
}

func TestClientCertTagsMTLS(t *testing.T) {
	pki := testutil.NewPKI("../../../testutil/pki")
	serverCfg := pki.TLSServerConfig()
	serverCfg.TLSCipherSuites = nil
	serverCfg.TLSMinVersion = ""
	serverCfg.TLSMaxVersion = ""

	plugin := &OpenTelemetry{
		ServerConfig:   *serverCfg,
		ServiceAddress: "127.0.0.1:0",
		ClientCertTags: []string{"common_name", "san_dns", "issuer_common_name"},
	}
	require.NoError(t, plugin.Init())

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	clientTLS, err := pki.TLSClientConfig().TLSConfig()
	require.NoError(t, err)
	clientTLS.ServerName = "localhost"

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(clientTLS)),
		otlpmetricgrpc.WithDialOption(
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
				return net.Dial("tcp", plugin.listener.Addr().String())
			})),
	)
	require.NoError(t, err)
	defer exporter.Shutdown(ctx) //nolint:errcheck // nothing to do on failure

	reader := sdkmetric.NewManualReader()
	defer reader.Shutdown(ctx) //nolint:errcheck // nothing to do on failure
	meter := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)).Meter("lib")
	counter, err := meter.Int64Counter("mtls-counter")
	require.NoError(t, err)
	counter.Add(ctx, 1)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &rm))
	require.NoError(t, exporter.Export(ctx, &rm))

	require.Empty(t, acc.Errors)
	actual := acc.GetTelegrafMetrics()
	require.Len(t, actual, 1)
	tags := actual[0].Tags()
	require.Equal(t, "localhost", tags["tls_client_common_name"])
	require.Equal(t, "localhost", tags["tls_client_san_dns"])
	require.Equal(t, "Telegraf Test CA", tags["tls_client_issuer_common_name"])
}
