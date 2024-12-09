package inputs_snmp_legacy_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/inputs_snmp_legacy" // register migration
)

func TestNoMigration(t *testing.T) {
	input := []byte(`
[[inputs.snmp_legacy]]
  address = "192.168.2.2:161"
`)

	output, n, err := config.ApplyMigrations(input)
	require.NoError(t, err)
	require.Empty(t, strings.TrimSpace(string(output)))
	require.Equal(t, uint64(1), n)
}
