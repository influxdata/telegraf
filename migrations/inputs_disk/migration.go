package inputs_disk

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
	if rawDeprecatedMountpoints, found := plugin["mountpoints"]; found {
		applied = true

		// Convert the options to the actual type
		deprecatedMountpoints, ok := rawDeprecatedMountpoints.([]interface{})
		if !ok {
			err := fmt.Errorf("unexpected type for deprecated 'mountpoints' option: %T", rawDeprecatedMountpoints)
			return nil, "", err
		}

		// Merge the option with the replacement
		var mountpoints []string
		if rawMountpoints, found := plugin["mount_points"]; found {
			mountpointsList, ok := rawMountpoints.([]interface{})
			if !ok {
				err := fmt.Errorf("unexpected type for 'mount_points' option: %T", rawMountpoints)
				return nil, "", err
			}
			for _, raw := range mountpointsList {
				mp, ok := raw.(string)
				if !ok {
					err := fmt.Errorf("unexpected type for 'mount_points' option: %T", raw)
					return nil, "", err
				}
				mountpoints = append(mountpoints, mp)
			}
		}
		for _, raw := range deprecatedMountpoints {
			dmp, ok := raw.(string)
			if !ok {
				err := fmt.Errorf("unexpected type for deprecated 'mountpoints' option: %T", raw)
				return nil, "", err
			}

			if !choice.Contains(dmp, mountpoints) {
				mountpoints = append(mountpoints, dmp)
			}
		}

		// Remove the deprecated option and replace the modified one
		delete(plugin, "mountpoints")
		plugin["mount_points"] = mountpoints
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "disk")
	cfg.Add("inputs", "disk", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.disk", migrate)
}
