package opentelemetry

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/testutil"
)

func TestOpenTelemetry(t *testing.T) {
	mockListener := bufconn.Listen(1024 * 1024)
	plugin := inputs.Inputs["opentelemetry"]().(*OpenTelemetry)
	plugin.listener = mockListener
	accumulator := new(testutil.Accumulator)

	err := plugin.Start(accumulator)
	require.NoError(t, err)
	t.Cleanup(plugin.Stop)

	metricExporter, err := otlpmetricgrpc.New(context.Background(),
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithDialOption(
			grpc.WithBlock(),
			grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
				return mockListener.Dial()
			})),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = metricExporter.Shutdown(context.Background()) })

	pusher := controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(),
			metricExporter,
		),
		controller.WithExporter(metricExporter),
	)

	err = pusher.Start(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { _ = pusher.Stop(context.Background()) })

	global.SetMeterProvider(pusher)

	// write metrics
	meter := global.MeterProvider().Meter("library-name")
	counter, err := meter.SyncInt64().Counter("measurement-counter")
	require.NoError(t, err)
	counter.Add(context.Background(), 7)

	err = pusher.Stop(context.Background())
	require.NoError(t, err)

	// Shutdown

	plugin.Stop()

	err = metricExporter.Shutdown(context.Background())
	require.NoError(t, err)

	// Check

	require.Empty(t, accumulator.Errors)

	require.Len(t, accumulator.Metrics, 1)
	got := accumulator.Metrics[0]
	require.Equal(t, "measurement-counter", got.Measurement)
	require.Equal(t, telegraf.Counter, got.Type)
	require.Equal(t, "library-name", got.Tags["otel.library.name"])
}
