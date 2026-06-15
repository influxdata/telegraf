package outputs_kafka

import (
	"fmt"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate the deprecated TLS options to their
// replacements.
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Check for deprecated options and migrate them
	var applied bool
	renames := []struct {
		from string
		to   string
	}{
		{"ca", "tls_ca"},
		{"certificate", "tls_cert"},
		{"key", "tls_key"},
	}
	for _, rename := range renames {
		rawOldValue, found := plugin[rename.from]
		if !found {
			continue
		}
		applied = true

		// Convert to the actual type
		oldValue, ok := rawOldValue.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for %q", rawOldValue, rename.from)
		}

		// Check if the new option already exists and if it has a contradicting
		// value. If the new option is not present, migrate the old value.
		if rawNewValue, found := plugin[rename.to]; !found {
			plugin[rename.to] = oldValue
		} else if newValue, ok := rawNewValue.(string); !ok {
			return nil, "", fmt.Errorf("unexpected type %T for %q", rawNewValue, rename.to)
		} else if newValue != oldValue {
			return nil, "", fmt.Errorf("contradicting setting for %q and %q", rename.to, rename.from)
		}

		// Remove the deprecated setting
		delete(plugin, rename.from)
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("outputs", "kafka")
	cfg.Add("outputs", "kafka", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("outputs.kafka", migrate)
}
