package opentelemetry

import (
	"context"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"google.golang.org/grpc/credentials/insecure"
	"time"

	"github.com/influxdata/influxdb-observability/common"
	"github.com/influxdata/influxdb-observability/influx2otel"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/outputs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	// This causes the gRPC library to register gzip compression.
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/metadata"
)

type OpenTelemetry struct {
	ServiceAddress string `toml:"service_address"`

	tls.ClientConfig
	Timeout     config.Duration   `toml:"timeout"`
	Compression string            `toml:"compression"`
	Headers     map[string]string `toml:"headers"`
	Attributes  map[string]string `toml:"attributes"`

	Log telegraf.Logger `toml:"-"`

	metricsConverter     *influx2otel.LineProtocolToOtelMetrics
	grpcClientConn       *grpc.ClientConn
	metricsServiceClient pmetricotlp.Client
	callOptions          []grpc.CallOption
}

func (o *OpenTelemetry) Connect() error {
	logger := &otelLogger{o.Log}

	if o.ServiceAddress == "" {
		o.ServiceAddress = defaultServiceAddress
	}
	if o.Timeout <= 0 {
		o.Timeout = defaultTimeout
	}
	if o.Compression == "" {
		o.Compression = defaultCompression
	}

	metricsConverter, err := influx2otel.NewLineProtocolToOtelMetrics(logger)
	if err != nil {
		return err
	}

	var grpcTLSDialOption grpc.DialOption
	if tlsConfig, err := o.ClientConfig.TLSConfig(); err != nil {
		return err
	} else if tlsConfig != nil {
		grpcTLSDialOption = grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig))
	} else {
		grpcTLSDialOption = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	grpcClientConn, err := grpc.Dial(o.ServiceAddress, grpcTLSDialOption)
	if err != nil {
		return err
	}

	metricsServiceClient := pmetricotlp.NewClient(grpcClientConn)

	o.metricsConverter = metricsConverter
	o.grpcClientConn = grpcClientConn
	o.metricsServiceClient = metricsServiceClient

	if o.Compression != "" && o.Compression != "none" {
		o.callOptions = append(o.callOptions, grpc.UseCompressor(o.Compression))
	}

	return nil
}

func (o *OpenTelemetry) Close() error {
	if o.grpcClientConn != nil {
		err := o.grpcClientConn.Close()
		o.grpcClientConn = nil
		return err
	}
	return nil
}

func (o *OpenTelemetry) Write(metrics []telegraf.Metric) error {
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
			o.Log.Warnf("unrecognized metric type %Q", metric.Type())
			continue
		}
		err := batch.AddPoint(metric.Name(), metric.Tags(), metric.Fields(), metric.Time(), vType)
		if err != nil {
			o.Log.Warnf("failed to add point: %s", err)
			continue
		}
	}

	md := pmetricotlp.NewRequest()
	md.SetMetrics(batch.GetMetrics())
	if md.Metrics().ResourceMetrics().Len() == 0 {
		return nil
	}

	if len(o.Attributes) > 0 {
		for i := 0; i < md.Metrics().ResourceMetrics().Len(); i++ {
			for k, v := range o.Attributes {
				md.Metrics().ResourceMetrics().At(i).Resource().Attributes().UpsertString(k, v)
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(o.Timeout))

	if len(o.Headers) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(o.Headers))
	}
	defer cancel()
	_, err := o.metricsServiceClient.Export(ctx, md, o.callOptions...)
	return err
}

const (
	defaultServiceAddress = "localhost:4317"
	defaultTimeout        = config.Duration(5 * time.Second)
	defaultCompression    = "gzip"
)

func init() {
	outputs.Add("opentelemetry", func() telegraf.Output {
		return &OpenTelemetry{
			ServiceAddress: defaultServiceAddress,
			Timeout:        defaultTimeout,
			Compression:    defaultCompression,
		}
	})
}
