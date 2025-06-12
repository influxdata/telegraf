package inputs_http_listener

import (
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate http_listener to influxdb_listener
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old plugin configuration
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Create the new plugin configuration with the same settings
	// but under the "influxdb_listener" plugin name instead of "http_listener"
	cfg := migrations.CreateTOMLStruct("inputs", "influxdb_listener")
	cfg.Add("inputs", "influxdb_listener", plugin)

	output, err := toml.Marshal(cfg)
	message := "migrated from deprecated 'http_listener' to 'influxdb_listener'"
	return output, message, err
}

// Register the migration function for the deprecated plugin
func init() {
	migrations.AddPluginMigration("inputs.http_listener", migrate)
}
