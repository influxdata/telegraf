package inputs_zookeeper

import (
	"errors"
	"fmt"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate the deprecated 'enable_tls' option to its
// replacement 'tls_enable'.
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Check for the deprecated option and migrate it
	var applied bool
	if rawOldValue, found := plugin["enable_tls"]; found {
		applied = true

		// Convert to the actual type
		oldValue, ok := rawOldValue.(bool)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'enable_tls'", rawOldValue)
		}

		// Check if the new option already exists and if it has a contradicting
		// value. If the new option is not present, migrate the old value.
		if rawNewValue, found := plugin["tls_enable"]; found {
			if newValue, ok := rawNewValue.(bool); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'tls_enable'", rawNewValue)
			} else if newValue != oldValue {
				return nil, "", errors.New("contradicting setting for 'tls_enable' and 'enable_tls'")
			}
		} else {
			plugin["tls_enable"] = oldValue
		}

		// Remove the deprecated setting
		delete(plugin, "enable_tls")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "zookeeper")
	cfg.Add("inputs", "zookeeper", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.zookeeper", migrate)
}
