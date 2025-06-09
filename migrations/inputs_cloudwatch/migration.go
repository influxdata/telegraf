package inputs_cloudwatch

import (
	"fmt"
	"strings"

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
	var messages []string

	// Check if deprecated 'namespace' option exists
	if namespaceValue, found := plugin["namespace"]; found {
		applied = true

		// Get the namespace value as string
		namespaceStr, ok := namespaceValue.(string)
		if !ok {
			return nil, "", fmt.Errorf("namespace value is not a string: %T", namespaceValue)
		}

		// Check if 'namespaces' already exists
		if namespacesValue, exists := plugin["namespaces"]; exists {
			// namespaces already exists, we need to merge them
			namespacesSlice, ok := namespacesValue.([]interface{})
			if !ok {
				return nil, "", fmt.Errorf("namespaces value is not a slice: %T", namespacesValue)
			}

			// Convert to string slice for easier handling
			var existingNamespaces []string
			for _, ns := range namespacesSlice {
				if nsStr, ok := ns.(string); ok {
					existingNamespaces = append(existingNamespaces, nsStr)
				}
			}

			// Check if the namespace is already in namespaces
			found := false
			for _, existing := range existingNamespaces {
				if existing == namespaceStr {
					found = true
					break
				}
			}

			if !found {
				// Add the namespace to the existing namespaces
				existingNamespaces = append(existingNamespaces, namespaceStr)
				// Convert back to []interface{} for TOML
				var newNamespaces []interface{}
				for _, ns := range existingNamespaces {
					newNamespaces = append(newNamespaces, ns)
				}
				plugin["namespaces"] = newNamespaces
				messages = append(messages, fmt.Sprintf("merged 'namespace' value '%s' into existing 'namespaces' array", namespaceStr))
			} else {
				messages = append(messages, fmt.Sprintf("removed deprecated 'namespace' option (value '%s' already exists in 'namespaces')", namespaceStr))
			}
		} else {
			// namespaces doesn't exist, create it with the namespace value
			plugin["namespaces"] = []interface{}{namespaceStr}
			messages = append(messages, fmt.Sprintf("migrated 'namespace' option to 'namespaces' array with value '%s'", namespaceStr))
		}

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
	message := strings.Join(messages, "; ")

	return output, message, err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.cloudwatch", migrate)
}
