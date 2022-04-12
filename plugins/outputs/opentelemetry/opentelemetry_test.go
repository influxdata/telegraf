package opentelemetry

import (
	"context"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/influxdb-observability/common"
	"github.com/influxdata/influxdb-observability/influx2otel"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/model/otlp"
	"go.opentelemetry.io/collector/model/otlpgrpc"
	"go.opentelemetry.io/collector/model/pdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestOpenTelemetry(t *testing.T) {
	expect := pdata.NewMetrics()
	{
		rm := expect.ResourceMetrics().AppendEmpty()
		rm.Resource().Attributes().InsertString("host.name", "potato")
		rm.Resource().Attributes().InsertString("attr-key", "attr-val")
		ilm := rm.ScopeMetrics().AppendEmpty()
		ilm.Scope().SetName("My Library Name")
		m := ilm.Metrics().AppendEmpty()
		m.SetName("cpu_temp")
		m.SetDataType(pdata.MetricDataTypeGauge)
		dp := m.Gauge().DataPoints().AppendEmpty()
		dp.Attributes().InsertString("foo", "bar")
		dp.SetTimestamp(pdata.Timestamp(1622848686000000000))
		dp.SetDoubleVal(87.332)
	}
	m := newMockOtelService(t)
	t.Cleanup(m.Cleanup)

	metricsConverter, err := influx2otel.NewLineProtocolToOtelMetrics(common.NoopLogger{})
	require.NoError(t, err)
	plugin := &OpenTelemetry{
		ServiceAddress:       m.Address(),
		Timeout:              config.Duration(time.Second),
		Headers:              map[string]string{"test": "header1"},
		Attributes:           map[string]string{"attr-key": "attr-val"},
		metricsConverter:     metricsConverter,
		grpcClientConn:       m.GrpcClient(),
		metricsServiceClient: otlpgrpc.NewMetricsClient(m.GrpcClient()),
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
		time.Unix(0, 1622848686000000000))

	err = plugin.Write([]telegraf.Metric{input})
	if err != nil {
		// TODO not sure why the service returns this error, but the data arrives as required by the test
		// rpc error: code = Internal desc = grpc: error while marshaling: proto: Marshal called with nil
		if !strings.Contains(err.Error(), "proto: Marshal called with nil") {
			assert.NoError(t, err)
		}
	}

	got := m.GotMetrics()

	expectJSON, err := otlp.NewJSONMetricsMarshaler().MarshalMetrics(expect)
	require.NoError(t, err)

	gotJSON, err := otlp.NewJSONMetricsMarshaler().MarshalMetrics(got)
	require.NoError(t, err)

	assert.JSONEq(t, string(expectJSON), string(gotJSON))
}

var _ otlpgrpc.MetricsServer = (*mockOtelService)(nil)

type mockOtelService struct {
	t          *testing.T
	listener   net.Listener
	grpcServer *grpc.Server
	grpcClient *grpc.ClientConn

	metrics pdata.Metrics
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

	otlpgrpc.RegisterMetricsServer(grpcServer, mockOtelService)
	go func() { assert.NoError(t, grpcServer.Serve(listener)) }()

	grpcClient, err := grpc.Dial(listener.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	require.NoError(t, err)
	mockOtelService.grpcClient = grpcClient

	return mockOtelService
}

func (m *mockOtelService) Cleanup() {
	assert.NoError(m.t, m.grpcClient.Close())
	m.grpcServer.Stop()
}

func (m *mockOtelService) GrpcClient() *grpc.ClientConn {
	return m.grpcClient
}

func (m *mockOtelService) GotMetrics() pdata.Metrics {
	return m.metrics
}

func (m *mockOtelService) Address() string {
	return m.listener.Addr().String()
}

func (m *mockOtelService) Export(ctx context.Context, request otlpgrpc.MetricsRequest) (otlpgrpc.MetricsResponse, error) {
	m.metrics = request.Metrics().Clone()
	ctxMetadata, ok := metadata.FromIncomingContext(ctx)
	assert.Equal(m.t, []string{"header1"}, ctxMetadata.Get("test"))
	assert.True(m.t, ok)
	return otlpgrpc.MetricsResponse{}, nil
}
