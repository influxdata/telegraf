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
	if rawOldUnits, found := plugin["supervisor_unit"]; found {
		applied = true

		// Check if the new option already exists and merge the two
		var units []string
		if newUnits, found := plugin["supervisor_units"]; found {
			var err error
			units, err = migrations.AsStringSlice(newUnits)
			if err != nil {
				return nil, "", fmt.Errorf("setting 'supervisor_units': %w", err)
			}
		}
		oldUnits, err := migrations.AsStringSlice(rawOldUnits)
		if err != nil {
			return nil, "", fmt.Errorf("setting 'supervisor_unit': %w", err)
		}
		for _, u := range oldUnits {
			if !choice.Contains(u, units) {
				units = append(units, u)
			}
		}
		plugin["supervisor_units"] = units

		// Remove deprecated setting
		delete(plugin, "supervisor_unit")
	}

	// The tagging options both need the 'tag_with' setting
	var tagwith []string
	if rawNewTagWith, found := plugin["tag_with"]; found {
		var err error
		tagwith, err = migrations.AsStringSlice(rawNewTagWith)
		if err != nil {
			return nil, "", fmt.Errorf("setting 'tag_with': %w", err)
		}
	}

	// Tagging with PID
	if oldPidTag, found := plugin["pid_tag"]; found {
		applied = true

		pt, ok := oldPidTag.(bool)
		if !ok {
			return nil, "", fmt.Errorf("setting 'pid_tag' has wrong type %T", oldPidTag)
		}

		// Add the pid-tagging to 'tag_with' if requested
		if pt && !choice.Contains("pid", tagwith) {
			tagwith = append(tagwith, "pid")
			plugin["tag_with"] = tagwith
		}

		// Remove deprecated setting
		delete(plugin, "pid_tag")
	}

	// Tagging with command-line
	if oldCmdlinedTag, found := plugin["cmdline_tag"]; found {
		applied = true

		ct, ok := oldCmdlinedTag.(bool)
		if !ok {
			return nil, "", fmt.Errorf("setting 'cmdline_tag' has wrong type %T", oldCmdlinedTag)
		}

		// Add the pid-tagging to 'tag_with' if requested
		if ct && !choice.Contains("cmdline", tagwith) {
			tagwith = append(tagwith, "cmdline")
			plugin["tag_with"] = tagwith
		}

		// Remove deprecated setting
		delete(plugin, "cmdline_tag")
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
