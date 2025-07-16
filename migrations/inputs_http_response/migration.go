package inputs_http_response

import (
	"fmt"
	"slices"

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
	if rawOldAddress, found := plugin["address"]; found {
		applied = true

		// Convert the options to the actual type
		oldAddress, ok := rawOldAddress.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'address'", rawOldAddress)
		}

		// Merge the option with the replacement
		var urls []string
		if rawNewURLs, found := plugin["urls"]; found {
			var err error
			urls, err = migrations.AsStringSlice(rawNewURLs)
			if err != nil {
				return nil, "", fmt.Errorf("'directories' option: %w", err)
			}
		}

		if !slices.Contains(urls, oldAddress) {
			urls = append(urls, oldAddress)
		}

		// Remove the deprecated option and replace the modified one
		plugin["urls"] = urls
		delete(plugin, "address")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "http_response")
	cfg.Add("inputs", "http_response", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.http_response", migrate)
}
