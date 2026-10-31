package kafka

import (
	"strings"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/metric"
	"github.com/influxdata/telegraf/models"
	"github.com/influxdata/telegraf/plugins/serializers/influx"
	"github.com/influxdata/telegraf/testutil"
)

func TestWriteConfigurationErrorsWithBroker(t *testing.T) {
	tests := []struct {
		name    string
		version string
		values  []string
		headers map[string]string
		accept  []int
		reject  []int
	}{
		{
			name:    "oversized message in mixed batch",
			version: "0.11.0.0",
			values:  []string{strings.Repeat("x", 1024), "small", "later"},
			accept:  []int{1},
			reject:  []int{0},
		},
		{
			name:    "oversized messages fill batch",
			version: "0.11.0.0",
			values:  []string{strings.Repeat("x", 1024), strings.Repeat("y", 1024), "later"},
			reject:  []int{0, 1},
		},
		{
			name:    "headers unsupported by broker version",
			version: "0.10.2.0",
			values:  []string{"first", "second"},
			headers: map[string]string{"name": "{{ .Name }}"},
			reject:  []int{0, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serializer := &influx.Serializer{}
			require.NoError(t, serializer.Init())
			plugin := &Kafka{
				Topic:        "existing",
				Headers:      tt.headers,
				Log:          testutil.Logger{},
				producerFunc: sarama.NewSyncProducer,
				serializer:   serializer,
			}
			plugin.Version = tt.version
			plugin.MaxMessageBytes = 256
			plugin.RequiredAcks = 1
			// Configure Sarama's global logger before the broker starts goroutines.
			require.NoError(t, plugin.Init())
			broker := sarama.NewMockBroker(t, 0)
			defer broker.Close()
			broker.SetHandlerByMap(map[string]sarama.MockResponse{
				"MetadataRequest": sarama.NewMockMetadataResponse(t).
					SetBroker(broker.Addr(), broker.BrokerID()).
					SetLeader("existing", 0, broker.BrokerID()),
				"ProduceRequest": sarama.NewMockProduceResponse(t),
			})
			plugin.Brokers = []string{broker.Addr()}
			plugin.saramaConfig.ApiVersionsRequest = false
			plugin.saramaConfig.Metadata.Retry.Max = 0
			plugin.saramaConfig.Net.DialTimeout = 2 * time.Second
			plugin.saramaConfig.Net.ReadTimeout = 2 * time.Second
			plugin.saramaConfig.Net.WriteTimeout = 2 * time.Second
			require.NoError(t, plugin.Connect())
			defer func() { require.NoError(t, plugin.Close()) }()

			buffer, err := models.NewBuffer("kafka", t.Name(), "", 10, "memory", "", false)
			require.NoError(t, err)
			defer func() { require.NoError(t, buffer.Close()) }()
			delivered := make(map[int]bool)
			for i, value := range tt.values {
				m, _ := metric.WithTracking(
					metric.New("cpu", nil, map[string]any{"value": value}, time.Unix(1, 0)),
					func(info telegraf.DeliveryInfo) { delivered[i] = info.Delivered() },
				)
				require.Zero(t, buffer.Add(m))
			}

			// Use Sarama's real client-side validation, not injected broker errors.
			tx := buffer.BeginTransaction(2)
			var partial *internal.PartialWriteError
			require.ErrorAs(t, plugin.Write(tx.Batch), &partial)
			require.Equal(t, tt.accept, partial.MetricsAccept)
			require.Equal(t, tt.reject, partial.MetricsReject)
			require.Len(t, partial.MetricsRejectErrors, len(tt.reject))
			for _, err := range partial.MetricsRejectErrors {
				var configErr sarama.ConfigurationError
				require.ErrorAs(t, err, &configErr)
			}
			tx.Accept, tx.Reject = partial.MetricsAccept, partial.MetricsReject
			buffer.EndTransaction(tx)
			require.Len(t, delivered, 2)
			for _, i := range tt.reject {
				require.False(t, delivered[i])
			}
			for _, i := range tt.accept {
				require.True(t, delivered[i])
			}

			// Even a full batch of permanent failures must let later metrics drain.
			if len(tt.values) > 2 {
				require.Equal(t, 1, buffer.Len())
				tx = buffer.BeginTransaction(2)
				require.NoError(t, plugin.Write(tx.Batch))
				tx.AcceptAll()
				buffer.EndTransaction(tx)
				require.Len(t, delivered, 3)
				require.True(t, delivered[2])
			}
			require.Zero(t, buffer.Len())
		})
	}
}
