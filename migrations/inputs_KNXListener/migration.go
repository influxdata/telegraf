package inputs_KNXListener

import (
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate KNXListener to knx_listener
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old plugin configuration
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Create the new plugin configuration with the same settings
	// but under the "knx_listener" plugin name instead of "KNXListener"
	cfg := migrations.CreateTOMLStruct("inputs", "knx_listener")
	cfg.Add("inputs", "knx_listener", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the deprecated plugin
func init() {
	migrations.AddPluginMigration("inputs.KNXListener", migrate)
}
