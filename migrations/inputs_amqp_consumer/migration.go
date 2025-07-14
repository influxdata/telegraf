package inputs_amqp_consumer

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
	if rawOldURL, found := plugin["url"]; found {
		applied = true

		// Convert the options to the actual type
		oldURL, ok := rawOldURL.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'url'", rawOldURL)
		}

		// Merge the option with the replacement
		var brokers []string
		if rawBrokers, found := plugin["brokers"]; found {
			var err error
			brokers, err = migrations.AsStringSlice(rawBrokers)
			if err != nil {
				return nil, "", fmt.Errorf("'brokers' option: %w", err)
			}
		}
		if !slices.Contains(brokers, oldURL) {
			brokers = append(brokers, oldURL)
		}

		// Remove the deprecated option and replace the modified one
		delete(plugin, "url")
		plugin["brokers"] = brokers
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "amqp_consumer")
	cfg.Add("inputs", "amqp_consumer", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.amqp_consumer", migrate)
}
