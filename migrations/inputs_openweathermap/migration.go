package inputs_openweathermap

import (
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to remove the deprecated 'query_style' option. The option
// is ignored at runtime since the upstream OpenWeatherMap v2.5 group API was
// removed, so it has no replacement and is simply dropped.
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Check for the deprecated option and remove it
	if _, found := plugin["query_style"]; !found {
		return nil, "", migrations.ErrNotApplicable
	}
	delete(plugin, "query_style")

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "openweathermap")
	cfg.Add("inputs", "openweathermap", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.openweathermap", migrate)
}
