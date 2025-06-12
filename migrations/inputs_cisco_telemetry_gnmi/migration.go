package inputs_cisco_telemetry_gnmi

import (
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate cisco_telemetry_gnmi to gnmi
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old plugin configuration
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Create the new plugin configuration with the same settings
	// but under the "gnmi" plugin name instead of "cisco_telemetry_gnmi"
	cfg := migrations.CreateTOMLStruct("inputs", "gnmi")
	cfg.Add("inputs", "gnmi", plugin)

	output, err := toml.Marshal(cfg)
	message := "migrated from deprecated 'cisco_telemetry_gnmi' to 'gnmi'"
	return output, message, err
}

// Register the migration function for the deprecated plugin
func init() {
	migrations.AddPluginMigration("inputs.cisco_telemetry_gnmi", migrate)
}
