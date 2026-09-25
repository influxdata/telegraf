package outputs_amon

import (
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

const msg = `
    The 'outputs.amon' plugin has been removed as the Amon service no longer
    exists. There is no replacement; please remove the plugin from your
    configuration.
`

// Migration function
func migrate(_ *ast.Table) ([]byte, string, error) {
	return nil, msg, nil
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginMigration("outputs.amon", migrate)
}
