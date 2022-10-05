package opentelemetry

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/influxdata/telegraf/testutil"
)

func TestOpenTelemetry(t *testing.T) {
	// create mock OpenTelemetry client

	mockListener := bufconn.Listen(1024 * 1024)
	t.Cleanup(func() { _ = mockListener.Close() })
	plugin := inputs.Inputs["opentelemetry"]().(*OpenTelemetry)
	plugin.listener = mockListener
	accumulator := new(testutil.Accumulator)

	require.NoError(t, plugin.Start(accumulator))
	t.Cleanup(plugin.Stop)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	metricExporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithDialOption(
			grpc.WithBlock(),
			grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
				return mockListener.DialContext(ctx)
			})),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = metricExporter.Shutdown(ctx) })

	reader := metric.NewManualReader()
	mp := metric.NewMeterProvider(metric.WithReader(reader))

	// set a metric value

	meter := mp.Meter("library-name")
	counter, err := meter.SyncInt64().Counter("measurement-counter")
	counter.Add(ctx, 7)

	// write metrics through the telegraf OpenTelemetry input plugin

	rm, err := reader.Collect(ctx)
	require.NoError(t, err)
	require.NoError(t, metricExporter.Export(ctx, rm))

	// Shutdown

	require.NoError(t, reader.Shutdown(ctx))
	require.NoError(t, metricExporter.Shutdown(ctx))
	plugin.Stop()

	// Check

	require.Empty(t, accumulator.Errors)
	require.Len(t, accumulator.Metrics, 1)
	got := accumulator.Metrics[0]
	require.Equal(t, "measurement-counter", got.Measurement)
	require.Equal(t, telegraf.Counter, got.Type)
	require.Equal(t, "library-name", got.Tags["otel.library.name"])
}
