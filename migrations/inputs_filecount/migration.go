package inputs_filecount

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
	if rawOldDirectory, found := plugin["directory"]; found {
		applied = true

		// Convert the options to the actual type
		oldDirectory, ok := rawOldDirectory.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'directory'", rawOldDirectory)
		}

		// Merge the option with the replacement
		var directories []string
		if rawNewDirectories, found := plugin["directories"]; found {
			var err error
			directories, err = migrations.AsStringSlice(rawNewDirectories)
			if err != nil {
				return nil, "", fmt.Errorf("'directories' option: %w", err)
			}
		}

		if !slices.Contains(directories, oldDirectory) {
			directories = append(directories, oldDirectory)
		}

		// Remove the deprecated option and replace the modified one
		delete(plugin, "directory")
		plugin["directories"] = directories
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "filecount")
	cfg.Add("inputs", "filecount", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.filecount", migrate)
}
