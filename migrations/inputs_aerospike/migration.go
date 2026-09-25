package inputs_aerospike

import (
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

const msg = `
    This plugin cannot be migrated automatically and requires manual intervention!
    Please use 'inputs.prometheus' with the Aerospike Prometheus Exporter instead.
`

// Migration function
func migrate(_ *ast.Table) ([]byte, string, error) {
	return nil, msg, nil
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginMigration("inputs.aerospike", migrate)
}
