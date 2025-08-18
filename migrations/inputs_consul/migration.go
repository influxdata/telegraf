package inputs_consul

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
	if rawOldDatacentre, found := plugin["datacentre"]; found {
		// Convert the options to the actual type
		oldDatacentre, ok := rawOldDatacentre.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'datacentre'", rawOldDatacentre)
		}

		// Check if the new setting is present and if so, check if the values are
		// conflicting.
		if rawNewDatacenter, found := plugin["datacenter"]; found {
			if newDatacenter, ok := rawNewDatacenter.(string); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'datacenter'", rawNewDatacenter)
			} else if oldDatacentre != newDatacenter {
				return nil, "", errors.New("contradicting setting for 'datacentre' and 'datacenter'")
			}
		}
		applied = true

		// Remove the deprecated option and replace the modified one
		plugin["datacenter"] = oldDatacentre
		delete(plugin, "datacentre")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "consul")
	cfg.Add("inputs", "consul", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.consul", migrate)
}
