package inputs_icinga2

import (
	"fmt"
	"slices"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Check for deprecated option(s) and migrate them
	var applied bool
	if rawOldObjectType, found := plugin["object_type"]; found {
		applied = true

		// Convert the options to the actual type
		oldObjectTypes, ok := rawOldObjectType.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'object_type'", rawOldObjectType)
		}

		// Merge the option with the replacement
		var objects []string
		if rawNewObjects, found := plugin["objects"]; found {
			var err error
			objects, err = migrations.AsStringSlice(rawNewObjects)
			if err != nil {
				return nil, "", fmt.Errorf("'objects' option: %w", err)
			}
		}

		if !slices.Contains(objects, oldObjectTypes) {
			objects = append(objects, oldObjectTypes)
		}

		// Remove the deprecated option and replace the modified one
		plugin["objects"] = objects
		delete(plugin, "object_type")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "icinga2")
	cfg.Add("inputs", "icinga2", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.icinga2", migrate)
}
