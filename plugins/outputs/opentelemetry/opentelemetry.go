//go:generate ../../../tools/readme_config_includer/generator
package opentelemetry

import (
	"context"
	ntls "crypto/tls"
	_ "embed"
	"fmt"
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
	ServiceAddress string `toml:"service_address"`
	Protocol       string `toml:"protocol"`
	EncodingType   string `toml:"encoding_type"`

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

type otlpMetricClient interface {
	Export(ctx context.Context, request pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error)
	Close() error
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
	o.otlpMetricClient.Close()
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
