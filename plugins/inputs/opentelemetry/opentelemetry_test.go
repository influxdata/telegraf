package opentelemetry

import (
	"context"
	"net"
	"testing"
	"time"

	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/influxdata/telegraf"
	tmetric "github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/testutil"
)

func TestOpenTelemetry(t *testing.T) {
	// Setup the plugin with a direct mockup connection
	listener := bufconn.Listen(1024 * 1024)
	defer listener.Close()

	plugin := &OpenTelemetry{
		MetricsSchema: "prometheus-v1",
		listener:      listener,
	}

	var acc testutil.Accumulator
	require.NoError(t, plugin.Start(&acc))
	defer plugin.Stop()

	// Setup the OpenTelemetry exporter
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)

	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithDialOption(
			grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
				return listener.DialContext(ctx)
			})),
	)
	require.NoError(t, err)
	defer exporter.Shutdown(ctx)

	// Setup the metric to send
	reader := metric.NewManualReader()
	defer reader.Shutdown(ctx)
	provider := metric.NewMeterProvider(metric.WithReader(reader))
	meter := provider.Meter("library-name")
	counter, err := meter.Int64Counter("measurement-counter")
	require.NoError(t, err)
	counter.Add(ctx, 7)

	// Write the OpenTelemetry metrics
	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(ctx, &rm))
	require.NoError(t, exporter.Export(ctx, &rm))

	// Shutdown
	require.NoError(t, reader.Shutdown(ctx))
	require.NoError(t, exporter.Shutdown(ctx))
	plugin.Stop()

	// Check
	require.Empty(t, acc.Errors)

	expected := []telegraf.Metric{
		tmetric.New(
			"measurement-counter",
			map[string]string{
				"otel.library.name":      "library-name",
				"service.name":           "unknown_service:opentelemetry.test",
				"telemetry.sdk.language": "go",
				"telemetry.sdk.name":     "opentelemetry",
				"telemetry.sdk.version":  "1.27.0",
			},
			map[string]interface{}{
				"counter": 7,
			},
			time.Unix(0, 0),
			telegraf.Counter,
		),
	}
	actual := acc.GetTelegrafMetrics()
	testutil.RequireMetricsEqual(t, expected, actual, testutil.IgnoreTime(), testutil.IgnoreFields("start_time_unix_nano"))
}
