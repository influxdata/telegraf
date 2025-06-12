package inputs_cloudwatch

import (
	"fmt"
	"slices"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate deprecated namespace option to namespaces
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	var applied bool

	// Check if deprecated 'namespace' option exists
	if namespaceValue, found := plugin["namespace"]; found {
		applied = true

		// Get the namespace value as string
		namespaceStr, ok := namespaceValue.(string)
		if !ok {
			return nil, "", fmt.Errorf("namespace value is not a string: %T", namespaceValue)
		}

		// Merge the option with the replacement
		var namespaces []string
		if rawNewNamespaces, found := plugin["namespaces"]; found {
			var err error
			namespaces, err = migrations.AsStringSlice(rawNewNamespaces)
			if err != nil {
				return nil, "", fmt.Errorf("'namespaces' option: %w", err)
			}
		}

		if !slices.Contains(namespaces, namespaceStr) {
			namespaces = append(namespaces, namespaceStr)
		}

		// Update the plugin configuration
		plugin["namespaces"] = namespaces

		// Remove the deprecated setting
		delete(plugin, "namespace")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "cloudwatch")
	cfg.Add("inputs", "cloudwatch", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.cloudwatch", migrate)
}
