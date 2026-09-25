package inputs_aerospike_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/inputs_aerospike" // register migration
)

func TestMigration(t *testing.T) {
	input := []byte(`
[[inputs.aerospike]]
  servers = ["localhost:3000"]
  username = "telegraf"
  password = "secret"
  enable_tls = true
  tls_name = "tlsname"
`)

	// The plugin cannot be migrated automatically, so it is dropped with a notice
	output, n, err := config.ApplyMigrations(input)
	require.NoError(t, err)
	require.Empty(t, strings.TrimSpace(string(output)))
	require.Equal(t, uint64(1), n)
}
