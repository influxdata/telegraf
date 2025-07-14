package inputs_gnmi

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
	if raw, found := plugin["guess_path_tag"]; found {
		applied = true

		if v, ok := raw.(bool); ok && v {
			plugin["path_guessing_strategy"] = "common path"
		}

		// Remove the ignored setting
		delete(plugin, "guess_path_tag")
	}

	if rawOldEnableTLS, found := plugin["enable_tls"]; found {
		// Convert the options to the actual type
		oldEnableTLS, ok := rawOldEnableTLS.(bool)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'enable_tls'", rawOldEnableTLS)
		}

		// Check if the new setting is present and if so, check if the values are
		// conflicting.
		if rawNewTLSEnable, found := plugin["tls_enable"]; found {
			if newTLSEnable, ok := rawNewTLSEnable.(bool); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'tls_enable'", rawNewTLSEnable)
			} else if oldEnableTLS != newTLSEnable {
				return nil, "", errors.New("contradicting setting for 'enable_tls' and 'tls_enable'")
			}
		}
		applied = true

		// Remove the deprecated option and replace the modified one
		plugin["tls_enable"] = oldEnableTLS
		delete(plugin, "enable_tls")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "gnmi")
	cfg.Add("inputs", "gnmi", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.gnmi", migrate)
}
