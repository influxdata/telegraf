package inputs_kafka_consumer_legacy_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/inputs_kafka_consumer_legacy" // register migration
)

func TestNoMigration(t *testing.T) {
	input := []byte(`
[[inputs.kafka_consumer_legacy]]
  topics = ["telegraf"]
  zookeeper_peers = ["localhost:2181"]
  zookeeper_chroot = ""
  consumer_group = "telegraf_metrics_consumers"
  offset = "oldest"
  data_format = "influx"
  max_message_len = 65536
`)

	output, n, err := config.ApplyMigrations(input)
	require.NoError(t, err)
	require.Empty(t, strings.TrimSpace(string(output)))
	require.Equal(t, uint64(1), n)
}
