package inputs_elasticsearch

import (
	"fmt"
	"strings"

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
	var messages []string

	// Check if deprecated 'http_timeout' option exists
	if httpTimeoutValue, found := plugin["http_timeout"]; found {
		applied = true

		// Check if 'timeout' already exists
		if timeoutValue, exists := plugin["timeout"]; exists {
			// Both exist - remove deprecated one but keep existing timeout
			messages = append(messages, fmt.Sprintf("removed deprecated 'http_timeout' option (kept existing 'timeout' value: %v)", timeoutValue))
		} else {
			// Only http_timeout exists - migrate it to timeout
			plugin["timeout"] = httpTimeoutValue
			messages = append(messages, fmt.Sprintf("migrated 'http_timeout' option to 'timeout' with value: %v", httpTimeoutValue))
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
	message := strings.Join(messages, "; ")

	return output, message, err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.elasticsearch", migrate)
}
