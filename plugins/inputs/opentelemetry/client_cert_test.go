package opentelemetry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
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
	plugin := &OpenTelemetry{ClientCertTags: []string{"cn", "foo"}}
	require.ErrorContains(t, plugin.Init(), `invalid certificate attribute "foo"`)
}

func TestClientCertTags(t *testing.T) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(255),
		Subject:      pkix.Name{CommonName: "client-1"},
		Issuer:       pkix.Name{CommonName: "My CA"},
		DNSNames:     []string{"a.example.org"},
	}

	// Tags are prefixed and only the requested ones are returned
	expected := map[string]string{
		"tls_client_cn":     "client-1",
		"tls_client_issuer": "My CA",
		"tls_client_sn":     "ff",
		"tls_client_dns":    "a.example.org",
	}
	require.Equal(t, expected, clientCertTags(peerContext(cert), []string{"cn", "issuer", "sn", "dns"}))
	require.Equal(t, map[string]string{"tls_client_cn": "client-1"}, clientCertTags(peerContext(cert), []string{"cn"}))

	// Nothing without options, peer, TLS or client certificate
	require.Nil(t, clientCertTags(peerContext(cert), nil))
	require.Nil(t, clientCertTags(context.Background(), []string{"cn"}))
	require.Nil(t, clientCertTags(peerContext(nil), []string{"cn"}))
	plain := peer.NewContext(context.Background(), &peer.Peer{})
	require.Nil(t, clientCertTags(plain, []string{"cn"}))
}

func TestClientCertTagsOverrideAttributes(t *testing.T) {
	cert := &x509.Certificate{Subject: pkix.Name{CommonName: "client-1"}}
	var acc testutil.Accumulator
	w := &writeToAccumulator{accumulator: &acc, clientCertTags: []string{"cn"}}

	// The tag derived from the certificate wins over the one sent by the client
	require.NoError(t, w.EnqueuePoint(
		peerContext(cert), "test",
		map[string]string{"tls_client_cn": "spoofed", "other": "x"},
		map[string]any{"value": 1}, time.Unix(0, 0), 0,
	))
	// Also works if the sender did not provide any tags
	require.NoError(t, w.EnqueuePoint(
		peerContext(cert), "test", nil,
		map[string]any{"value": 2}, time.Unix(0, 0), 0,
	))

	metrics := acc.GetTelegrafMetrics()
	require.Len(t, metrics, 2)
	require.Equal(t, map[string]string{"tls_client_cn": "client-1", "other": "x"}, metrics[0].Tags())
	require.Equal(t, map[string]string{"tls_client_cn": "client-1"}, metrics[1].Tags())
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
		ClientCertTags: []string{"cn", "dns", "issuer"},
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
	require.Equal(t, "localhost", tags["tls_client_cn"])
	require.Equal(t, "localhost", tags["tls_client_dns"])
	require.Equal(t, "Telegraf Test CA", tags["tls_client_issuer"])
}
