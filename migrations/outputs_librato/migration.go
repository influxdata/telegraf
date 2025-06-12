package outputs_librato

import (
	"errors"
	"fmt"

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
	if rawSourceTag, found := plugin["source_tag"]; found {
		// Convert the options to the actual type
		sourceTag, ok := rawSourceTag.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'source_tag'", rawSourceTag)
		}

		if rawTemplate, found := plugin["template"]; found {
			if template, ok := rawTemplate.(string); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'template'", rawTemplate)
			} else if sourceTag != template {
				return nil, "", errors.New("contradicting setting for 'source_tag' and 'template'")
			}
		}

		applied = true
		plugin["template"] = sourceTag
		delete(plugin, "source_tag")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("outputs", "librato")
	cfg.Add("outputs", "librato", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("outputs.librato", migrate)
}
