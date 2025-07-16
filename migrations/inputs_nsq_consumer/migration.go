package inputs_nsq_consumer

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
	if rawOldServer, found := plugin["server"]; found {
		applied = true

		// Convert the options to the actual type
		oldServer, ok := rawOldServer.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'server'", rawOldServer)
		}

		// Merge the option with the replacement
		var nsqd []string
		if rawNewNSQD, found := plugin["nsqd"]; found {
			var err error
			nsqd, err = migrations.AsStringSlice(rawNewNSQD)
			if err != nil {
				return nil, "", fmt.Errorf("'nsqd' option: %w", err)
			}
		}

		if !slices.Contains(nsqd, oldServer) {
			nsqd = append(nsqd, oldServer)
		}

		// Remove the deprecated option and replace the modified one
		plugin["nsqd"] = nsqd
		delete(plugin, "server")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "nsq_consumer")
	cfg.Add("inputs", "nsq_consumer", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.nsq_consumer", migrate)
}
