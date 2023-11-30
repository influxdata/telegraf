package inputs_procstat

import (
	"fmt"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/internal/choice"
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
	if oldUnits, found := plugin["supervisor_unit"]; found {
		applied = true

		// Check if the new option already exists and merge the two
		var units []string
		if newUnits, found := plugin["supervisor_units"]; found {
			nu, ok := newUnits.([]interface{})
			if !ok {
				return nil, "", fmt.Errorf("setting 'supervisor_units' has wrong type %T", newUnits)
			}
			for _, raw := range nu {
				u, ok := raw.(string)
				if !ok {
					return nil, "", fmt.Errorf("setting 'supervisor_units' contains wrong type %T", raw)
				}
				units = append(units, u)
			}
		}
		ou, ok := oldUnits.([]interface{})
		if !ok {
			return nil, "", fmt.Errorf("setting 'supervisor_unit' has wrong type %T", oldUnits)
		}
		for _, raw := range ou {
			u, ok := raw.(string)
			if !ok {
				return nil, "", fmt.Errorf("setting 'supervisor_unit' contains wrong type %T", raw)
			}
			if !choice.Contains(u, units) {
				units = append(units, u)
			}
		}
		plugin["supervisor_units"] = units

		// Remove deprecated setting
		delete(plugin, "supervisor_unit")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "procstat")
	cfg.Add("inputs", "procstat", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.procstat", migrate)
}
