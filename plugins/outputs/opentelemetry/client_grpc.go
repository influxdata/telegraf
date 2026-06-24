package opentelemetry

import (
	"context"
	ntls "crypto/tls"
	"fmt"
	"net"

	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type gRPCClient struct {
	grpcClientConn       *grpc.ClientConn
	metricsServiceClient pmetricotlp.GRPCClient
	callOptions          []grpc.CallOption
}

func (g *gRPCClient) Connect(cfg *clientConfig) error {
	var grpcTLSDialOption grpc.DialOption
	if tlsConfig, err := cfg.TLSConfig.TLSConfig(); err != nil {
		return err
	} else if tlsConfig != nil {
		grpcTLSDialOption = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	} else if cfg.CoralogixConfig != nil {
		// For coralogix, we enforce GRPC connection with TLS
		grpcTLSDialOption = grpc.WithTransportCredentials(credentials.NewTLS(&ntls.Config{}))
	} else {
		grpcTLSDialOption = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	dialer, err := cfg.TCPProxy.Proxy()
	if err != nil {
		return fmt.Errorf("creating proxy failed: %w", err)
	}

	grpcDialOption := grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, "tcp", addr)
	})

	grpcClientConn, err := grpc.NewClient(cfg.ServiceAddress, grpcTLSDialOption, grpcDialOption, grpc.WithUserAgent(userAgent))
	if err != nil {
		return err
	}

	g.grpcClientConn = grpcClientConn
	g.metricsServiceClient = pmetricotlp.NewGRPCClient(grpcClientConn)

	if cfg.Compression != "" && cfg.Compression != "none" {
		g.callOptions = append(g.callOptions, grpc.UseCompressor(cfg.Compression))
	}

	return nil
}

func (g *gRPCClient) Export(ctx context.Context, request pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error) {
	return g.metricsServiceClient.Export(ctx, request, g.callOptions...)
}

func (g *gRPCClient) Close() error {
	if g == nil {
		return nil
	}

	if g.grpcClientConn != nil {
		err := g.grpcClientConn.Close()
		g.grpcClientConn = nil
		return err
	}
	return nil
}
