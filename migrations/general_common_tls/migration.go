package general_common_tls

import (
	"fmt"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

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

	// Option mapping to replace
	mapping := map[string]string{
		"ssl_ca":   "tls_ca",
		"ssl_cert": "tls_cert",
		"ssl_key":  "tls_key",
	}

	// Check if the old settings are present and set the new TLS settings. Warn
	// on conflicting settings where both the old and the new setting is
	// present.
	for oldSetting, newSetting := range mapping {
		rawOld, found := plugin[oldSetting]
		if !found {
			continue
		}
		vOld, ok := rawOld.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for %q", rawOld, oldSetting)
		}
		rawNew, present := plugin[newSetting]
		if present {
			if vNew, ok := rawNew.(string); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for %q", rawOld, newSetting)
			} else if vOld != vNew {
				return nil, "", fmt.Errorf("contradicting setting for %q and %q", oldSetting, newSetting)
			}
		} else {
			plugin[newSetting] = vOld
		}
		applied = true

		// Remove the deprecated option
		delete(plugin, oldSetting)
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
