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
		deprecatedMountpoints, err := migrations.AsStringSlice(rawDeprecatedMountpoints)
		if err != nil {
			return nil, "", fmt.Errorf("'mountpoints' option: %w", err)
		}

		// Merge the option with the replacement
		var mountpoints []string
		if rawMountpoints, found := plugin["mount_points"]; found {
			mountpoints, err = migrations.AsStringSlice(rawMountpoints)
			if err != nil {
				return nil, "", fmt.Errorf("'mount_points' option: %w", err)
			}
		}
		for _, dmp := range deprecatedMountpoints {
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
