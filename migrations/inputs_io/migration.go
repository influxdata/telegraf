package inputs_io

import (
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Load the plugin config and directly encode it again as we do not want to
	// modify anything beyond the name.

	// Decode the old data structure
	var plugin interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Create the corresponding metric configurations
	cfg := migrations.CreateTOMLStruct("inputs", "diskio")
	cfg.Add("inputs", "diskio", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginMigration("inputs.io", migrate)
}
