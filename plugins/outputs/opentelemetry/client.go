package opentelemetry

import (
	"bytes"
	"context"
	ntls "crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
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

type httpClient struct {
	httpClient   *http.Client
	url          string
	encodingType string
	compress     string
}

func (g *gRPCClient) Connect(
	serviceAddress string,
	clientConfig *tls.ClientConfig,
	compression string,
	coralogixConfig *CoralogixConfig,
	encoding string,
) error {
	gRPCClient := &gRPCClient{}
	var grpcTLSDialOption grpc.DialOption
	if tlsConfig, err := clientConfig.TLSConfig(); err != nil {
		return err
	} else if tlsConfig != nil {
		grpcTLSDialOption = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	} else if coralogixConfig != nil {
		// For coralogix, we enforce GRPC connection with TLS
		grpcTLSDialOption = grpc.WithTransportCredentials(credentials.NewTLS(&ntls.Config{}))
	} else {
		grpcTLSDialOption = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	grpcClientConn, err := grpc.NewClient(serviceAddress, grpcTLSDialOption, grpc.WithUserAgent(userAgent))
	if err != nil {
		return err
	}

	metricsServiceClient := pmetricotlp.NewGRPCClient(grpcClientConn)

	gRPCClient.grpcClientConn = grpcClientConn
	gRPCClient.metricsServiceClient = metricsServiceClient

	if compression != "" && compression != "none" {
		gRPCClient.callOptions = append(gRPCClient.callOptions, grpc.UseCompressor(compression))
	}

	return nil
}

func (h *httpClient) Connect(
	serviceAddress string,
	clientConfig *tls.ClientConfig,
	compression string,
	coralogixConfig *CoralogixConfig,
	encoding string,
) error {
	httpClient := &httpClient{
		httpClient:   &http.Client{},
		url:          serviceAddress,
		encodingType: encoding,
		compress:     compression,
	}
	if tlsConfig, err := clientConfig.TLSConfig(); err != nil {
		return err
	} else if tlsConfig != nil {
		httpClient.httpClient.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	} else if coralogixConfig != nil {
		// For coralogix, we enforce HTTP connection with TLS
		httpClient.httpClient.Transport = &http.Transport{
			TLSClientConfig: &ntls.Config{},
		}
	} else {
		httpClient.httpClient.Transport = &http.Transport{
			TLSClientConfig: &ntls.Config{
				InsecureSkipVerify: true,
			},
		}
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

func (h *httpClient) Export(ctx context.Context, request pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error) {
	var err error
	var requestBytes []byte

	switch h.encodingType {
	case "application/x-protobuf":
		requestBytes, err = request.MarshalProto()
		if err != nil {
			return pmetricotlp.ExportResponse{}, err
		}
	case "application/json":
		requestBytes, err = request.MarshalJSON()
		if err != nil {
			return pmetricotlp.ExportResponse{}, err
		}
	default:
		return pmetricotlp.ExportResponse{}, fmt.Errorf("unsupported content type '%s'", h.encodingType)
	}
	var reader io.Reader
	reader = bytes.NewReader(requestBytes)

	if h.compress != "" && h.compress != "none" {
		reader = internal.CompressWithGzip(reader)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, h.url, reader)
	if err != nil {
		return pmetricotlp.ExportResponse{}, err
	}

	httpRequest.Header.Set("Content-Type", h.encodingType)
	httpRequest.Header.Set("User-Agent", userAgent)
	if h.compress != "" && h.compress != "none" {
		httpRequest.Header.Set("Content-Encoding", "gzip")
	}

	httpResponse, err := h.httpClient.Do(httpRequest)
	if err != nil {
		return pmetricotlp.ExportResponse{}, err
	}
	defer func() {
		//nolint:errcheck // cannot fail with io.Discard
		io.CopyN(io.Discard, httpResponse.Body, maxHTTPResponseReadBytes)
		_ = httpResponse.Body.Close()
	}()

	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		return pmetricotlp.ExportResponse{}, fmt.Errorf("received non-2xx HTTP response code: %d", httpResponse.StatusCode)
	}

	responseBytes, err := readResponseBody(httpResponse)
	if err != nil {
		return pmetricotlp.ExportResponse{}, err
	}

	exportResponse := pmetricotlp.NewExportResponse()
	switch httpResponse.Header.Get("Content-Type") {
	case "application/x-protobuf":
		err = exportResponse.UnmarshalProto(responseBytes)
		if err != nil {
			return pmetricotlp.ExportResponse{}, err
		}
	case "application/json":
		err = exportResponse.UnmarshalJSON(responseBytes)
		if err != nil {
			return pmetricotlp.ExportResponse{}, err
		}
	}

	return exportResponse, nil
}

func (*httpClient) Close() error {
	// No persistent connections to close for HTTP client
	return nil
}

// ref. https://github.com/open-telemetry/opentelemetry-collector/blob/7258150320ae4c3b489aa58bd2939ba358b23ae1/exporter/otlphttpexporter/otlp.go#L271
func readResponseBody(resp *http.Response) ([]byte, error) {
	if resp.ContentLength == 0 {
		return nil, nil
	}

	maxRead := resp.ContentLength

	// if maxRead == -1, the ContentLength header has not been sent, so read up to
	// the maximum permitted body size. If it is larger than the permitted body
	// size, still try to read from the body in case the value is an error. If the
	// body is larger than the maximum size, proto unmarshaling will likely fail.
	if maxRead == -1 || maxRead > maxHTTPResponseReadBytes {
		maxRead = maxHTTPResponseReadBytes
	}
	protoBytes := make([]byte, maxRead)
	n, err := io.ReadFull(resp.Body, protoBytes)

	// No bytes read and an EOF error indicates there is no body to read.
	if n == 0 && (err == nil || errors.Is(err, io.EOF)) {
		return nil, nil
	}

	// io.ReadFull will return io.ErrorUnexpectedEOF if the Content-Length header
	// wasn't set, since we will try to read past the length of the body. If this
	// is the case, the body will still have the full message in it, so we want to
	// ignore the error and parse the message.
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}

	return protoBytes[:n], nil
}

const (
	maxHTTPResponseReadBytes = 64 * 1024 // 64 KB
)
