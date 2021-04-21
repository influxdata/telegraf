package opentelemetry

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/config"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
	metricsService "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type metricServiceServer struct {
	status *status.Status
	metricsService.UnimplementedMetricsServiceServer
	reqs []*metricsService.ExportMetricsServiceRequest
}

func (s *metricServiceServer) Export(ctx context.Context, req *metricsService.ExportMetricsServiceRequest) (*metricsService.ExportMetricsServiceResponse, error) {
	var emptyValue = metricsService.ExportMetricsServiceResponse{}
	s.reqs = append(s.reqs, req)

	if s.status == nil {
		return &emptyValue, nil
	}

	return nil, s.status.Err()
}

func (s *metricServiceServer) clear() {
	s.reqs = []*metricsService.ExportMetricsServiceRequest{}
}

func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("httptest: failed to listen on a port: %v", err))
		}
	}
	return l
}

var (
	listener          net.Listener
	mockMetricsServer metricServiceServer
)

func TestMain(m *testing.M) {
	listener = newLocalListener()
	grpcServer := grpc.NewServer()
	mockMetricsServer = metricServiceServer{
		status: nil,
	}
	metricsService.RegisterMetricsServiceServer(grpcServer, &mockMetricsServer)
	go func() {
		_ = grpcServer.Serve(listener)
	}()
	defer grpcServer.Stop()
	os.Exit(m.Run())
}
func TestConfigOptions(t *testing.T) {
	o := OpenTelemetry{
		Endpoint: ":::::",
		Log:      testutil.Logger{},
	}
	err := o.Init()
	require.Error(t, err)
	require.True(t, strings.HasPrefix(err.Error(), "invalid endpoint configured"))

	o = OpenTelemetry{
		Endpoint:    "http://" + listener.Addr().String(),
		Compression: "none",
		Log:         testutil.Logger{},
		ClientConfig: tls.ClientConfig{
			TLSCA: "invalid_ca",
		},
	}
	err = o.Init()
	require.Error(t, err)
	require.True(t, strings.HasPrefix(err.Error(), "invalid tls configuration"))

	o = OpenTelemetry{
		Endpoint:    "http://" + listener.Addr().String(),
		Compression: "none",
		Log:         testutil.Logger{},
	}
	err = o.Init()
	require.NoError(t, err)

	err = o.Connect()
	require.NoError(t, err)

	require.Equal(t, defaultTimeout, time.Duration(o.Timeout))
	require.Equal(t, map[string]string{"telemetry-reporting-agent": fmt.Sprint(
		"telegraf/",
		internal.Version(),
	)}, o.Headers)

	attributes := map[string]string{
		"service.name":    "test",
		"service.version": "0.0.1",
	}
	o = OpenTelemetry{
		Endpoint:    "http://" + listener.Addr().String(),
		Timeout:     config.Duration(time.Second * 10),
		Compression: "none",
		Attributes:  attributes,
		Log:         testutil.Logger{},
	}
	err = o.Init()
	require.NoError(t, err)

	err = o.Connect()
	require.NoError(t, err)

	require.Equal(t, time.Second*10, time.Duration(o.Timeout))
	require.Equal(t, len(o.resourceTags), 2)
	for _, tag := range o.resourceTags {
		require.Equal(t, attributes[tag.Key], tag.Value)
	}
}

func TestWrite(t *testing.T) {
	o := OpenTelemetry{
		Endpoint:    "http://" + listener.Addr().String(),
		Timeout:     config.Duration(time.Second * 10),
		Compression: "none",
		Log:         testutil.Logger{},
	}
	err := o.Init()
	require.NoError(t, err)

	err = o.Connect()
	require.NoError(t, err)

	mockMetricsServer.clear()
	err = o.Write(testutil.MockMetrics())
	require.NoError(t, err)

	require.Equal(t, 1, len(mockMetricsServer.reqs))
	request := mockMetricsServer.reqs[0]

	require.Equal(t, 1, len(request.ResourceMetrics[0].GetInstrumentationLibraryMetrics()))
	require.Equal(t, instrumentationLibraryName, request.ResourceMetrics[0].GetInstrumentationLibraryMetrics()[0].GetInstrumentationLibrary().GetName())
}

func TestWriteSupportedMetricKinds(t *testing.T) {
	// Metrics in descending order of timestamp
	metrics := []telegraf.Metric{
		testutil.MustMetric("cpu",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 42,
			},
			time.Unix(2, 0),
		),
		testutil.MustMetric("cpu",
			map[string]string{
				"foo": "foo",
			},
			map[string]interface{}{
				"value": 43,
			},
			time.Unix(3, 0),
		),
		testutil.MustMetric("ram",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 42.1,
			},
			time.Unix(4, 0),
		),
		testutil.MustMetric("up",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": false,
			},
			time.Unix(4, 0),
		),
		testutil.MustMetric("processes",
			map[string]string{
				"foo": "foo",
			},
			map[string]interface{}{
				"value": 43,
			},
			time.Unix(3, 0),
		),
		testutil.MustMetric("disk",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 43.9,
			},
			time.Unix(1, 0),
		),
	}
	o := OpenTelemetry{
		Endpoint:    "http://" + listener.Addr().String(),
		Timeout:     config.Duration(time.Second * 10),
		Compression: "none",
		Log:         testutil.Logger{},
	}
	err := o.Init()
	require.NoError(t, err)

	err = o.Connect()
	require.NoError(t, err)

	mockMetricsServer.clear()
	err = o.Write(metrics)
	require.NoError(t, err)

	require.Equal(t, 1, len(mockMetricsServer.reqs))
	require.Equal(t, len(metrics), len(mockMetricsServer.reqs[0].GetResourceMetrics()))
}

func TestWriteIgnoresInvalidKinds(t *testing.T) {
	// Metrics in descending order of timestamp
	metrics := []telegraf.Metric{
		testutil.MustMetric("custom_string_metric",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": "string value",
			},
			time.Unix(2, 0),
		),
		testutil.MustMetric("histogram",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 1,
			},
			time.Unix(2, 0),
			telegraf.Histogram,
		),
		testutil.MustMetric("summary",
			map[string]string{
				"foo": "bar",
			},
			map[string]interface{}{
				"value": 1,
			},
			time.Unix(2, 0),
			telegraf.Summary,
		),
	}
	o := OpenTelemetry{
		Endpoint:    "http://" + listener.Addr().String(),
		Timeout:     config.Duration(time.Second * 10),
		Compression: "none",
		Log:         testutil.Logger{},
	}
	err := o.Init()
	require.NoError(t, err)

	err = o.Connect()
	require.NoError(t, err)

	mockMetricsServer.clear()
	err = o.Write(metrics)
	require.NoError(t, err)

	require.Equal(t, 0, len(mockMetricsServer.reqs))
}
