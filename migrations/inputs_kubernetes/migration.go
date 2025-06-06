package inputs_kubernetes

import (
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate deprecated Kubernetes bearer_token_string option
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Check for deprecated bearer_token_string option
	var applied bool
	var message string

	if _, found := plugin["bearer_token_string"]; found {
		applied = true

		// Only migrate if bearer_token is not already set (don't overwrite existing)
		if _, bearerTokenExists := plugin["bearer_token"]; !bearerTokenExists {
			message = "removed deprecated 'bearer_token_string' option; please save the token to a file and use 'bearer_token' option with the file path instead"
		} else {
			message = "removed deprecated 'bearer_token_string' option; existing 'bearer_token' configuration preserved"
		}

		// Always remove the deprecated setting
		delete(plugin, "bearer_token_string")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "kubernetes")
	cfg.Add("inputs", "kubernetes", plugin)

	output, err := toml.Marshal(cfg)
	return output, message, err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.kubernetes", migrate)
}
