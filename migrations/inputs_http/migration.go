package inputs_http

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
	if rawOldBearerToken, found := plugin["bearer_token"]; found {
		// Convert the options to the actual type
		oldBearerToken, ok := rawOldBearerToken.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'bearer_token'", rawOldBearerToken)
		}

		// Check if the new setting is present and if so, check if the values are
		// conflicting.
		if rawNewTokenFile, found := plugin["token_file"]; found {
			if newTokenFile, ok := rawNewTokenFile.(string); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'token_file'", rawNewTokenFile)
			} else if oldBearerToken != newTokenFile {
				return nil, "", errors.New("contradicting setting for 'bearer_token' and 'token_file'")
			}
		}
		applied = true

		// Remove the deprecated option and replace the modified one
		plugin["token_file"] = oldBearerToken
		delete(plugin, "bearer_token")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "http")
	cfg.Add("inputs", "http", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.http", migrate)
}
