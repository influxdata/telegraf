package outputs_amon_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/outputs_amon" // register migration
)

func TestMigration(t *testing.T) {
	input := []byte(`
[[outputs.amon]]
  server_key = "my-server-key"
  amon_instance = "https://youramoninstance"
  timeout = "5s"
`)

	// The plugin has no replacement, so the migration drops it
	output, n, err := config.ApplyMigrations(input)
	require.NoError(t, err)
	require.Empty(t, strings.TrimSpace(string(output)))
	require.Equal(t, uint64(1), n)
}
