package outputs_remotefile

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
	if rawTrace, found := plugin["trace"]; found {
		// Convert the options to the actual type
		trace, ok := rawTrace.(bool)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'trace'", rawTrace)
		}

		if rawLogLevel, found := plugin["log_level"]; found {
			if logLevel, ok := rawLogLevel.(string); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'log_level'", rawLogLevel)
			} else if (trace && logLevel != "trace") || (!trace && logLevel == "trace") {
				return nil, "", errors.New("contradicting setting for 'trace' and 'log_level'")
			}
		}

		applied = true
		if trace {
			plugin["log_level"] = "trace"
		}
		delete(plugin, "trace")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("outputs", "remotefile")
	cfg.Add("outputs", "remotefile", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("outputs.remotefile", migrate)
}
