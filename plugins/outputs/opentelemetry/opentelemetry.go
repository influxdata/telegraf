//go:generate ../../../tools/readme_config_includer/generator
package opentelemetry

import (
	"bytes"
	"context"
	ntls "crypto/tls"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/influxdata/influxdb-observability/common"
	"github.com/influxdata/influxdb-observability/influx2otel"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/encoding/gzip" // Blank import to allow gzip encoding
	"google.golang.org/grpc/metadata"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
)

var userAgent = internal.ProductToken()

//go:embed sample.conf
var sampleConfig string

type OpenTelemetry struct {
	// Protocol is the protocol to use for sending data to the OpenTelemetry Collector.
	// Supported protocols are "grpc" & "http". Defaults to "grpc".
	Protocol string `toml:"protocol"`
	// EncodingType is the encoding type to use for sending data to the OpenTelemetry Collector.
	// It is used only when Protocol is set to "http".
	// Supported encoding types are "application/x-protobuf" & "application/json". Defaults to "application/x-protobuf".
	EncodingType string `toml:"encoding_type"`
	// ServiceAddress is the address of the OpenTelemetry Collector.
	// It must include the port number.
	// Example: "localhost:4317" for gRPC protocol, "http://localhost:4318/v1/metrics" for HTTP protocol.
	ServiceAddress string `toml:"service_address"`

	tls.ClientConfig
	Timeout     config.Duration   `toml:"timeout"`
	Compression string            `toml:"compression"`
	Headers     map[string]string `toml:"headers"`
	Attributes  map[string]string `toml:"attributes"`
	Coralogix   *CoralogixConfig  `toml:"coralogix"`

	Log telegraf.Logger `toml:"-"`

	metricsConverter *influx2otel.LineProtocolToOtelMetrics
	otlpMetricClient otlpMetricClient
}

var (
	_ otlpMetricClient = (*gRPCClient)(nil)
	_ otlpMetricClient = (*httpClient)(nil)
)

type otlpMetricClient interface {
	Export(ctx context.Context, request pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error)
	Close() error
}

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

type CoralogixConfig struct {
	AppName    string `toml:"application"`
	SubSystem  string `toml:"subsystem"`
	PrivateKey string `toml:"private_key"`
}

func (*OpenTelemetry) SampleConfig() string {
	return sampleConfig
}

func (o *OpenTelemetry) Connect() error {
	logger := &otelLogger{o.Log}
	if o.Protocol == "" {
		o.Protocol = defaultProtocol
	}
	if o.ServiceAddress == "" {
		o.ServiceAddress = defaultServiceAddress
	}
	if o.EncodingType == "" {
		o.EncodingType = defaultEncodingType
	}
	if o.Timeout <= 0 {
		o.Timeout = defaultTimeout
	}
	if o.Compression == "" {
		o.Compression = defaultCompression
	}
	if o.Coralogix != nil {
		if o.Headers == nil {
			o.Headers = make(map[string]string)
		}
		o.Headers["ApplicationName"] = o.Coralogix.AppName
		o.Headers["ApiName"] = o.Coralogix.SubSystem
		o.Headers["Authorization"] = "Bearer " + o.Coralogix.PrivateKey
	}

	metricsConverter, err := influx2otel.NewLineProtocolToOtelMetrics(logger)
	if err != nil {
		return err
	}
	o.metricsConverter = metricsConverter

	switch o.Protocol {
	case "", "grpc":
		err = o.connectGRPC()
		if err != nil {
			return err
		}
	case "http":
		err = o.connectHTTP()
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unsupported protocol '%s'", o.Protocol)
	}

	return nil
}

func (o *OpenTelemetry) Close() error {
	if o.otlpMetricClient != nil {
		return o.otlpMetricClient.Close()
	}
	return nil
}

func (o *OpenTelemetry) Write(metrics []telegraf.Metric) error {
	metricBatch := make(map[int64][]telegraf.Metric)
	timestamps := make([]int64, 0, len(metrics))
	for _, metric := range metrics {
		timestamp := metric.Time().UnixNano()
		if existingSlice, ok := metricBatch[timestamp]; ok {
			metricBatch[timestamp] = append(existingSlice, metric)
		} else {
			metricBatch[timestamp] = []telegraf.Metric{metric}
			timestamps = append(timestamps, timestamp)
		}
	}

	// sort the timestamps we collected
	sort.Slice(timestamps, func(i, j int) bool { return timestamps[i] < timestamps[j] })

	o.Log.Debugf("Received %d metrics and split into %d groups by timestamp", len(metrics), len(metricBatch))
	for _, timestamp := range timestamps {
		if err := o.sendBatch(metricBatch[timestamp]); err != nil {
			return err
		}
	}

	return nil
}

func (o *OpenTelemetry) sendBatch(metrics []telegraf.Metric) error {
	batch := o.metricsConverter.NewBatch()
	for _, metric := range metrics {
		var vType common.InfluxMetricValueType
		switch metric.Type() {
		case telegraf.Gauge:
			vType = common.InfluxMetricValueTypeGauge
		case telegraf.Untyped:
			vType = common.InfluxMetricValueTypeUntyped
		case telegraf.Counter:
			vType = common.InfluxMetricValueTypeSum
		case telegraf.Histogram:
			vType = common.InfluxMetricValueTypeHistogram
		case telegraf.Summary:
			vType = common.InfluxMetricValueTypeSummary
		default:
			o.Log.Warnf("Unrecognized metric type %v", metric.Type())
			continue
		}
		err := batch.AddPoint(metric.Name(), metric.Tags(), metric.Fields(), metric.Time(), vType)
		if err != nil {
			o.Log.Warnf("Failed to add point: %v", err)
			continue
		}
	}

	md := pmetricotlp.NewExportRequestFromMetrics(batch.GetMetrics())
	if md.Metrics().ResourceMetrics().Len() == 0 {
		return nil
	}

	if len(o.Attributes) > 0 {
		for i := 0; i < md.Metrics().ResourceMetrics().Len(); i++ {
			for k, v := range o.Attributes {
				md.Metrics().ResourceMetrics().At(i).Resource().Attributes().PutStr(k, v)
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Timeout))

	if len(o.Headers) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(o.Headers))
	}
	defer cancel()
	_, err := o.otlpMetricClient.Export(ctx, md)
	return err
}

func (o *OpenTelemetry) connectGRPC() error {
	gRPCClient := &gRPCClient{}
	var grpcTLSDialOption grpc.DialOption
	if tlsConfig, err := o.ClientConfig.TLSConfig(); err != nil {
		return err
	} else if tlsConfig != nil {
		grpcTLSDialOption = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	} else if o.Coralogix != nil {
		// For coralogix, we enforce GRPC connection with TLS
		grpcTLSDialOption = grpc.WithTransportCredentials(credentials.NewTLS(&ntls.Config{}))
	} else {
		grpcTLSDialOption = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	grpcClientConn, err := grpc.NewClient(o.ServiceAddress, grpcTLSDialOption, grpc.WithUserAgent(userAgent))
	if err != nil {
		return err
	}

	metricsServiceClient := pmetricotlp.NewGRPCClient(grpcClientConn)

	gRPCClient.grpcClientConn = grpcClientConn
	gRPCClient.metricsServiceClient = metricsServiceClient

	if o.Compression != "" && o.Compression != "none" {
		gRPCClient.callOptions = append(gRPCClient.callOptions, grpc.UseCompressor(o.Compression))
	}

	o.otlpMetricClient = gRPCClient
	return nil
}

func (o *OpenTelemetry) connectHTTP() error {
	httpClient := &httpClient{
		httpClient:   &http.Client{},
		url:          o.ServiceAddress,
		encodingType: o.EncodingType,
		compress:     o.Compression,
	}
	if tlsConfig, err := o.ClientConfig.TLSConfig(); err != nil {
		return err
	} else if tlsConfig != nil {
		httpClient.httpClient.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	} else if o.Coralogix != nil {
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

	o.otlpMetricClient = httpClient
	return nil
}

func (g *gRPCClient) Export(ctx context.Context, request pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error) {
	return g.metricsServiceClient.Export(ctx, request, g.callOptions...)
}

func (g *gRPCClient) Close() error {
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

const (
	defaultProtocol       = "grpc"
	defaultEncodingType   = "application/x-protobuf"
	defaultServiceAddress = "localhost:4317"
	defaultTimeout        = config.Duration(5 * time.Second)
	defaultCompression    = "gzip"
)

func init() {
	outputs.Add("opentelemetry", func() telegraf.Output {
		return &OpenTelemetry{
			Protocol:       defaultProtocol,
			EncodingType:   defaultEncodingType,
			ServiceAddress: defaultServiceAddress,
			Timeout:        defaultTimeout,
			Compression:    defaultCompression,
		}
	})
}
