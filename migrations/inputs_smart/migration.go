package inputs_smart

import (
	"errors"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

func migrate(tbl *ast.Table) ([]byte, string, error) {
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	var applied bool

	if path, found := plugin["path"]; found {
		if _, found := plugin["path_smartctl"]; found {
			return nil, "", errors.New("cannot migrate 'path' option, as 'path_smartctl' is already set")
		}
		plugin["path_smartctl"] = path
		delete(plugin, "path")
		applied = true
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "smart")
	cfg.Add("inputs", "smart", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.smart", migrate)
}
