package inputs_internet_speed

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
	if rawOldEnableFileDownload, found := plugin["enable_file_download"]; found {
		applied = true

		// Convert the options to the actual type
		oldEnableFileDownload, ok := rawOldEnableFileDownload.(bool)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'enable_file_download'", rawOldEnableFileDownload)
		}

		// Merge the option with the replacement
		if rawNewMemorySavingMode, found := plugin["memory_saving_mode"]; found {
			if newMemorySavingMode, ok := rawNewMemorySavingMode.(bool); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'memory_saving_mode'", rawNewMemorySavingMode)
			} else if newMemorySavingMode != oldEnableFileDownload {
				return nil, "", errors.New("contradicting setting for 'enable_file_download' and 'memory_saving_mode'")
			}
		} else if oldEnableFileDownload {
			plugin["memory_saving_mode"] = true
		}

		// Remove the deprecated option and replace the modified one
		delete(plugin, "enable_file_download")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "internet_speed")
	cfg.Add("inputs", "internet_speed", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.internet_speed", migrate)
}
