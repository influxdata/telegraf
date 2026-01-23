package opentelemetry

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/influxdata/influxdb-observability/common"
	"github.com/influxdata/influxdb-observability/influx2otel"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
)

func TestOpenTelemetry(t *testing.T) {
	expect := pmetric.NewMetrics()
	{
		rm := expect.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr("host.name", "potato")
		rm.Resource().Attributes().PutStr("attr-key", "attr-val")
		ilm := rm.ScopeMetrics().AppendEmpty()
		ilm.Scope().SetName("My Library Name")
		m := ilm.Metrics().AppendEmpty()
		m.SetName("cpu_temp")
		m.SetEmptyGauge()
		dp := m.Gauge().DataPoints().AppendEmpty()
		dp.Attributes().PutStr("foo", "bar")
		dp.SetTimestamp(pcommon.Timestamp(1622848686000000000))
		dp.SetDoubleValue(87.332)
	}
	m := newMockOtelService(t)
	t.Cleanup(m.Cleanup)

	metricsConverter, err := influx2otel.NewLineProtocolToOtelMetrics(common.NoopLogger{})
	require.NoError(t, err)
	plugin := &OpenTelemetry{
		ServiceAddress:   m.Address(),
		Timeout:          config.Duration(time.Second),
		Headers:          map[string]string{"test": "header1"},
		Attributes:       map[string]string{"attr-key": "attr-val"},
		metricsConverter: metricsConverter,
		otlpMetricClient: &gRPCClient{
			grpcClientConn:       m.GrpcClient(),
			metricsServiceClient: pmetricotlp.NewGRPCClient(m.GrpcClient()),
		},
		Log: testutil.Logger{},
	}

	input := testutil.MustMetric(
		"cpu_temp",
		map[string]string{
			"foo":               "bar",
			"otel.library.name": "My Library Name",
			"host.name":         "potato",
		},
		map[string]interface{}{
			"gauge": 87.332,
		},
		time.Unix(0, 1622848686000000000),
	)

	require.NoError(t, plugin.Write([]telegraf.Metric{input}))

	got := m.GotMetrics()

	marshaller := pmetric.JSONMarshaler{}
	expectJSON, err := marshaller.MarshalMetrics(expect)
	require.NoError(t, err)

	gotJSON, err := marshaller.MarshalMetrics(got)
	require.NoError(t, err)

	require.JSONEq(t, string(expectJSON), string(gotJSON))
}

func TestOpenTelemetryHTTPProtobuf(t *testing.T) {
	expect := pmetric.NewMetrics()
	{
		rm := expect.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr("host.name", "potato")
		rm.Resource().Attributes().PutStr("attr-key", "attr-val")
		ilm := rm.ScopeMetrics().AppendEmpty()
		ilm.Scope().SetName("My Library Name")
		m := ilm.Metrics().AppendEmpty()
		m.SetName("cpu_temp")
		m.SetEmptyGauge()
		dp := m.Gauge().DataPoints().AppendEmpty()
		dp.Attributes().PutStr("foo", "bar")
		dp.SetTimestamp(pcommon.Timestamp(1622848686000000000))
		dp.SetDoubleValue(87.332)
	}

	var receivedMetrics pmetric.Metrics
	var receivedContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		req := pmetricotlp.NewExportRequest()
		err = req.UnmarshalProto(body)
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		receivedMetrics = pmetric.NewMetrics()
		req.Metrics().CopyTo(receivedMetrics)

		resp := pmetricotlp.NewExportResponse()
		respBytes, err := resp.MarshalProto()
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(respBytes)
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	metricsConverter, err := influx2otel.NewLineProtocolToOtelMetrics(common.NoopLogger{})
	require.NoError(t, err)

	plugin := &OpenTelemetry{
		ServiceAddress:   server.URL,
		EncodingType:     "protobuf",
		Timeout:          config.Duration(time.Second),
		Attributes:       map[string]string{"attr-key": "attr-val"},
		Compression:      "none",
		metricsConverter: metricsConverter,
		Log:              testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())

	input := testutil.MustMetric(
		"cpu_temp",
		map[string]string{
			"foo":               "bar",
			"otel.library.name": "My Library Name",
			"host.name":         "potato",
		},
		map[string]interface{}{
			"gauge": 87.332,
		},
		time.Unix(0, 1622848686000000000),
	)

	require.NoError(t, plugin.Write([]telegraf.Metric{input}))

	require.Equal(t, "application/x-protobuf", receivedContentType)

	marshaller := pmetric.JSONMarshaler{}
	expectJSON, err := marshaller.MarshalMetrics(expect)
	require.NoError(t, err)

	gotJSON, err := marshaller.MarshalMetrics(receivedMetrics)
	require.NoError(t, err)

	require.JSONEq(t, string(expectJSON), string(gotJSON))
}

func TestOpenTelemetryHTTPJSON(t *testing.T) {
	expect := pmetric.NewMetrics()
	{
		rm := expect.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().PutStr("host.name", "potato")
		rm.Resource().Attributes().PutStr("attr-key", "attr-val")
		ilm := rm.ScopeMetrics().AppendEmpty()
		ilm.Scope().SetName("My Library Name")
		m := ilm.Metrics().AppendEmpty()
		m.SetName("cpu_temp")
		m.SetEmptyGauge()
		dp := m.Gauge().DataPoints().AppendEmpty()
		dp.Attributes().PutStr("foo", "bar")
		dp.SetTimestamp(pcommon.Timestamp(1622848686000000000))
		dp.SetDoubleValue(87.332)
	}

	var receivedMetrics pmetric.Metrics
	var receivedContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}

		req := pmetricotlp.NewExportRequest()
		err = req.UnmarshalJSON(body)
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		receivedMetrics = pmetric.NewMetrics()
		req.Metrics().CopyTo(receivedMetrics)

		resp := pmetricotlp.NewExportResponse()
		respBytes, err := resp.MarshalJSON()
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(respBytes)
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	metricsConverter, err := influx2otel.NewLineProtocolToOtelMetrics(common.NoopLogger{})
	require.NoError(t, err)

	plugin := &OpenTelemetry{
		ServiceAddress:   server.URL,
		EncodingType:     "json",
		Timeout:          config.Duration(time.Second),
		Attributes:       map[string]string{"attr-key": "attr-val"},
		Compression:      "none",
		metricsConverter: metricsConverter,
		Log:              testutil.Logger{},
	}
	require.NoError(t, plugin.Connect())

	input := testutil.MustMetric(
		"cpu_temp",
		map[string]string{
			"foo":               "bar",
			"otel.library.name": "My Library Name",
			"host.name":         "potato",
		},
		map[string]interface{}{
			"gauge": 87.332,
		},
		time.Unix(0, 1622848686000000000),
	)

	require.NoError(t, plugin.Write([]telegraf.Metric{input}))

	require.Equal(t, "application/json", receivedContentType)

	marshaller := pmetric.JSONMarshaler{}
	expectJSON, err := marshaller.MarshalMetrics(expect)
	require.NoError(t, err)

	gotJSON, err := marshaller.MarshalMetrics(receivedMetrics)
	require.NoError(t, err)

	require.JSONEq(t, string(expectJSON), string(gotJSON))
}

var _ pmetricotlp.GRPCServer = (*mockOtelService)(nil)

type mockOtelService struct {
	pmetricotlp.UnimplementedGRPCServer
	t          *testing.T
	listener   net.Listener
	grpcServer *grpc.Server
	grpcClient *grpc.ClientConn

	metrics pmetric.Metrics
}

func newMockOtelService(t *testing.T) *mockOtelService {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	grpcServer := grpc.NewServer()

	mockOtelService := &mockOtelService{
		t:          t,
		listener:   listener,
		grpcServer: grpcServer,
	}

	pmetricotlp.RegisterGRPCServer(grpcServer, mockOtelService)
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			t.Error(err)
		}
	}()

	grpcClient, err := grpc.NewClient(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	require.True(t, grpcClient.WaitForStateChange(t.Context(), connectivity.Connecting))
	mockOtelService.grpcClient = grpcClient

	return mockOtelService
}

func (m *mockOtelService) Cleanup() {
	require.NoError(m.t, m.grpcClient.Close())
	m.grpcServer.Stop()
}

func (m *mockOtelService) GrpcClient() *grpc.ClientConn {
	return m.grpcClient
}

func (m *mockOtelService) GotMetrics() pmetric.Metrics {
	return m.metrics
}

func (m *mockOtelService) Address() string {
	return m.listener.Addr().String()
}

func (m *mockOtelService) Export(ctx context.Context, request pmetricotlp.ExportRequest) (pmetricotlp.ExportResponse, error) {
	m.metrics = pmetric.NewMetrics()
	request.Metrics().CopyTo(m.metrics)
	// Only check metadata if it exists (for tests that provide headers)
	ctxMetadata, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if testHeader := ctxMetadata.Get("test"); len(testHeader) > 0 {
			require.Equal(m.t, []string{"header1"}, testHeader)
		}
	}
	return pmetricotlp.NewExportResponse(), nil
}

func TestOpenTelemetryMetricNameFormatPrometheus(t *testing.T) {
	expect := pmetric.NewMetrics()
	{
		rm := expect.ResourceMetrics().AppendEmpty()
		ilm := rm.ScopeMetrics().AppendEmpty()
		m := ilm.Metrics().AppendEmpty()
		m.SetName("http_server_duration") // dots converted to underscores
		m.SetEmptyGauge()
		dp := m.Gauge().DataPoints().AppendEmpty()
		dp.SetTimestamp(pcommon.Timestamp(1622848686000000000))
		dp.SetDoubleValue(87.332)
	}
	m := newMockOtelService(t)
	t.Cleanup(m.Cleanup)

	metricsConverter, err := influx2otel.NewLineProtocolToOtelMetrics(common.NoopLogger{})
	require.NoError(t, err)
	plugin := &OpenTelemetry{
		ServiceAddress:    m.Address(),
		Timeout:          config.Duration(time.Second),
		MetricNameFormat: "prometheus",
		metricsConverter: metricsConverter,
		otlpMetricClient: &gRPCClient{
			grpcClientConn:       m.GrpcClient(),
			metricsServiceClient: pmetricotlp.NewGRPCClient(m.GrpcClient()),
		},
		Log: testutil.Logger{},
	}

	input := testutil.MustMetric(
		"http.server.duration", // dot-separated name
		map[string]string{},
		map[string]interface{}{
			"gauge": 87.332,
		},
		time.Unix(0, 1622848686000000000),
	)

	require.NoError(t, plugin.Write([]telegraf.Metric{input}))

	got := m.GotMetrics()
	require.Equal(t, 1, got.ResourceMetrics().Len())
	require.Equal(t, 1, got.ResourceMetrics().At(0).ScopeMetrics().Len())
	require.Equal(t, 1, got.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().Len())
	require.Equal(t, "http_server_duration", got.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Name())
}

func TestOpenTelemetryMetricNameFormatOtel(t *testing.T) {
	expect := pmetric.NewMetrics()
	{
		rm := expect.ResourceMetrics().AppendEmpty()
		ilm := rm.ScopeMetrics().AppendEmpty()
		m := ilm.Metrics().AppendEmpty()
		m.SetName("http.server.duration") // dots preserved
		m.SetEmptyGauge()
		dp := m.Gauge().DataPoints().AppendEmpty()
		dp.SetTimestamp(pcommon.Timestamp(1622848686000000000))
		dp.SetDoubleValue(87.332)
	}
	m := newMockOtelService(t)
	t.Cleanup(m.Cleanup)

	metricsConverter, err := influx2otel.NewLineProtocolToOtelMetrics(common.NoopLogger{})
	require.NoError(t, err)
	plugin := &OpenTelemetry{
		ServiceAddress:    m.Address(),
		Timeout:          config.Duration(time.Second),
		MetricNameFormat: "otel",
		metricsConverter: metricsConverter,
		otlpMetricClient: &gRPCClient{
			grpcClientConn:       m.GrpcClient(),
			metricsServiceClient: pmetricotlp.NewGRPCClient(m.GrpcClient()),
		},
		Log: testutil.Logger{},
	}

	input := testutil.MustMetric(
		"http.server.duration", // dot-separated name
		map[string]string{},
		map[string]interface{}{
			"gauge": 87.332,
		},
		time.Unix(0, 1622848686000000000),
	)

	require.NoError(t, plugin.Write([]telegraf.Metric{input}))

	got := m.GotMetrics()
	require.Equal(t, 1, got.ResourceMetrics().Len())
	require.Equal(t, 1, got.ResourceMetrics().At(0).ScopeMetrics().Len())
	require.Equal(t, 1, got.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().Len())
	require.Equal(t, "http.server.duration", got.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Name())
}

func TestOpenTelemetryMetricNameFormatDefault(t *testing.T) {
	expect := pmetric.NewMetrics()
	{
		rm := expect.ResourceMetrics().AppendEmpty()
		ilm := rm.ScopeMetrics().AppendEmpty()
		m := ilm.Metrics().AppendEmpty()
		m.SetName("http_server_duration") // default is prometheus format
		m.SetEmptyGauge()
		dp := m.Gauge().DataPoints().AppendEmpty()
		dp.SetTimestamp(pcommon.Timestamp(1622848686000000000))
		dp.SetDoubleValue(87.332)
	}
	m := newMockOtelService(t)
	t.Cleanup(m.Cleanup)

	metricsConverter, err := influx2otel.NewLineProtocolToOtelMetrics(common.NoopLogger{})
	require.NoError(t, err)
	plugin := &OpenTelemetry{
		ServiceAddress:    m.Address(),
		Timeout:          config.Duration(time.Second),
		MetricNameFormat: "", // empty should default to prometheus
		metricsConverter: metricsConverter,
		otlpMetricClient: &gRPCClient{
			grpcClientConn:       m.GrpcClient(),
			metricsServiceClient: pmetricotlp.NewGRPCClient(m.GrpcClient()),
		},
		Log: testutil.Logger{},
	}
	require.NoError(t, plugin.Connect()) // Connect sets default

	input := testutil.MustMetric(
		"http.server.duration", // dot-separated name
		map[string]string{},
		map[string]interface{}{
			"gauge": 87.332,
		},
		time.Unix(0, 1622848686000000000),
	)

	require.NoError(t, plugin.Write([]telegraf.Metric{input}))

	got := m.GotMetrics()
	require.Equal(t, 1, got.ResourceMetrics().Len())
	require.Equal(t, 1, got.ResourceMetrics().At(0).ScopeMetrics().Len())
	require.Equal(t, 1, got.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().Len())
	require.Equal(t, "http_server_duration", got.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Name())
}

func TestOpenTelemetryInvalidMetricNameFormat(t *testing.T) {
	plugin := &OpenTelemetry{
		ServiceAddress:    "localhost:4317",
		MetricNameFormat: "invalid",
		Log:              testutil.Logger{},
	}
	err := plugin.Connect()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid metric_name_format")
}

func TestOpenTelemetryMetricNameWithUnderscores(t *testing.T) {
	// Test that existing underscores are preserved in both formats
	m := newMockOtelService(t)
	t.Cleanup(m.Cleanup)

	metricsConverter, err := influx2otel.NewLineProtocolToOtelMetrics(common.NoopLogger{})
	require.NoError(t, err)
	
	// Test prometheus format - underscores should remain, dots should be converted
	plugin := &OpenTelemetry{
		ServiceAddress:    m.Address(),
		Timeout:          config.Duration(time.Second),
		MetricNameFormat: "prometheus",
		metricsConverter: metricsConverter,
		otlpMetricClient: &gRPCClient{
			grpcClientConn:       m.GrpcClient(),
			metricsServiceClient: pmetricotlp.NewGRPCClient(m.GrpcClient()),
		},
		Log: testutil.Logger{},
	}

	input := testutil.MustMetric(
		"http.server.request.duration", // dots and underscores
		map[string]string{},
		map[string]interface{}{
			"gauge": 87.332,
		},
		time.Unix(0, 1622848686000000000),
	)

	require.NoError(t, plugin.Write([]telegraf.Metric{input}))
	got := m.GotMetrics()
	require.Equal(t, 1, got.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().Len())
	// Dots should be converted to underscores
	require.Equal(t, "http_server_request_duration", got.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Name())
	
	// Test otel format - everything should be preserved
	plugin2 := &OpenTelemetry{
		ServiceAddress:    m.Address(),
		Timeout:          config.Duration(time.Second),
		MetricNameFormat: "otel",
		metricsConverter: metricsConverter,
		otlpMetricClient: &gRPCClient{
			grpcClientConn:       m.GrpcClient(),
			metricsServiceClient: pmetricotlp.NewGRPCClient(m.GrpcClient()),
		},
		Log: testutil.Logger{},
	}

	input2 := testutil.MustMetric(
		"http.server.request.duration",
		map[string]string{},
		map[string]interface{}{
			"gauge": 87.332,
		},
		time.Unix(0, 1622848686000000000),
	)

	require.NoError(t, plugin2.Write([]telegraf.Metric{input2}))
	got2 := m.GotMetrics()
	require.Equal(t, 1, got2.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().Len())
	// Dots should be preserved
	require.Equal(t, "http.server.request.duration", got2.ResourceMetrics().At(0).ScopeMetrics().At(0).Metrics().At(0).Name())
}
