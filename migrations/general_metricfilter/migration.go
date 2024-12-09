package general_metricfilter

import (
	"fmt"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/internal/choice"
	"github.com/influxdata/telegraf/migrations"
)

// Migration function
func migrate(category, name string, tbl *ast.Table) ([]byte, string, error) {
	// Filter options can only be present in inputs, outputs, processors and
	// aggregators. Skip everything else...
	switch category {
	case "inputs", "outputs", "processors", "aggregators":
	default:
		return nil, "", migrations.ErrNotApplicable
	}

	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Check for deprecated option(s) and migrate them
	var applied bool

	// Get the new field settings to be able to merge it with the deprecated
	// settings
	var fieldinclude []string
	if newFieldInclude, found := plugin["fieldinclude"]; found {
		var err error
		fieldinclude, err = migrations.AsStringSlice(newFieldInclude)
		if err != nil {
			return nil, "", fmt.Errorf("setting 'fieldinclude': %w", err)
		}
	}
	for _, option := range []string{"pass", "fieldpass"} {
		if rawOld, found := plugin[option]; found {
			applied = true

			old, err := migrations.AsStringSlice(rawOld)
			if err != nil {
				return nil, "", fmt.Errorf("setting '%s': %w", option, err)
			}
			for _, o := range old {
				if !choice.Contains(o, fieldinclude) {
					fieldinclude = append(fieldinclude, o)
				}
			}

			// Remove the deprecated setting
			delete(plugin, option)
		}
	}
	// Add the new option if it has data
	if len(fieldinclude) > 0 {
		plugin["fieldinclude"] = fieldinclude
	}

	var fieldexclude []string
	if newFieldExclude, found := plugin["fieldexclude"]; found {
		var err error
		fieldexclude, err = migrations.AsStringSlice(newFieldExclude)
		if err != nil {
			return nil, "", fmt.Errorf("setting 'fieldexclude': %w", err)
		}
	}
	for _, option := range []string{"drop", "fielddrop"} {
		if rawOld, found := plugin[option]; found {
			applied = true

			old, err := migrations.AsStringSlice(rawOld)
			if err != nil {
				return nil, "", fmt.Errorf("setting '%s': %w", option, err)
			}
			for _, o := range old {
				if !choice.Contains(o, fieldexclude) {
					fieldexclude = append(fieldexclude, o)
				}
			}

			// Remove the deprecated setting
			delete(plugin, option)
		}
	}
	// Add the new option if it has data
	if len(fieldexclude) > 0 {
		plugin["fieldexclude"] = fieldexclude
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct(category, name)
	cfg.Add(category, name, plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddGeneralMigration(migrate)
}
