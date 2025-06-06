package inputs_openldap

import (
	"errors"
	"fmt"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate deprecated SSL options to TLS options
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Check for deprecated options and migrate them
	var applied bool

	// Migrate ssl -> tls
	if rawOldValue, found := plugin["ssl"]; found {
		applied = true

		// Convert to the actual type
		oldValue, ok := rawOldValue.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'ssl'", rawOldValue)
		}

		// Check if the new option already exists and if it has a contradicting
		// value. If the new option is not present, migrate the old value.
		if rawNewValue, found := plugin["tls"]; !found {
			plugin["tls"] = oldValue
		} else {
			if newValue, ok := rawNewValue.(string); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'tls'", rawNewValue)
			} else if newValue != oldValue {
				return nil, "", errors.New("contradicting setting for 'tls' and 'ssl'")
			}
		}

		// Remove the deprecated setting
		delete(plugin, "ssl")
	}

	// Migrate ssl_ca -> tls_ca
	if rawOldValue, found := plugin["ssl_ca"]; found {
		applied = true

		// Convert to the actual type
		oldValue, ok := rawOldValue.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'ssl_ca'", rawOldValue)
		}

		// Check if the new option already exists and if it has a contradicting
		// value. If the new option is not present, migrate the old value.
		if rawNewValue, found := plugin["tls_ca"]; !found {
			plugin["tls_ca"] = oldValue
		} else {
			if newValue, ok := rawNewValue.(string); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'tltls_cas'", rawNewValue)
			} else if newValue != oldValue {
				return nil, "", errors.New("contradicting setting for 'tls_ca' and 'ssl_ca'")
			}
		}

		// Remove the deprecated setting
		delete(plugin, "ssl_ca")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "openldap")
	cfg.Add("inputs", "openldap", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.openldap", migrate)
}
