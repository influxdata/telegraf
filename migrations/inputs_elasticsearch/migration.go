package inputs_elasticsearch

import (
	"fmt"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate a deprecated http_timeout option to timeout
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	var applied bool

	// Check if deprecated 'http_timeout' option exists
	if httpTimeoutValue, found := plugin["http_timeout"]; found {
		applied = true

		// Check if 'timeout' already exists
		if timeoutValue, exists := plugin["timeout"]; exists {
			// Both exist - check for conflicts
			if httpTimeoutValue != timeoutValue {
				return nil, "", fmt.Errorf("contradicting setting for 'http_timeout' (%v) and 'timeout' (%v)", httpTimeoutValue, timeoutValue)
			}
			// Values are the same, just remove the deprecated one
		} else {
			// Only http_timeout exists - migrate it to timeout
			plugin["timeout"] = httpTimeoutValue
		}

		// Remove the deprecated setting
		delete(plugin, "http_timeout")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "elasticsearch")
	cfg.Add("inputs", "elasticsearch", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.elasticsearch", migrate)
}
