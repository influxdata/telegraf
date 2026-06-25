package inputs_kafka_consumer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate the deprecated 'connection_strategy' option to
// its replacement 'startup_error_behavior'. The "defer" strategy, which allowed
// the plugin to start even if the broker was unavailable, maps to the "retry"
// behavior. The "startup" strategy (and the empty default) corresponds to the
// default "error" behavior, so the option is simply removed in that case.
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Check for the deprecated option and migrate it
	var applied bool
	if rawOldValue, found := plugin["connection_strategy"]; found {
		applied = true

		// Convert to the actual type
		oldValue, ok := rawOldValue.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'connection_strategy'", rawOldValue)
		}

		switch strings.ToLower(oldValue) {
		case "", "startup":
			// Default behavior corresponding to 'startup_error_behavior = "error"'
		case "defer":
			// Check if the new option already exists and if it has a
			// contradicting value. If the new option is not present, migrate
			// the old value to the equivalent behavior.
			if rawNewValue, found := plugin["startup_error_behavior"]; found {
				if newValue, ok := rawNewValue.(string); !ok {
					return nil, "", fmt.Errorf("unexpected type %T for 'startup_error_behavior'", rawNewValue)
				} else if newValue != "retry" {
					return nil, "", errors.New("contradicting setting for 'startup_error_behavior' and 'connection_strategy'")
				}
			} else {
				plugin["startup_error_behavior"] = "retry"
			}
		default:
			return nil, "", fmt.Errorf("invalid connection strategy %q", oldValue)
		}

		// Remove the deprecated setting
		delete(plugin, "connection_strategy")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "kafka_consumer")
	cfg.Add("inputs", "kafka_consumer", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.kafka_consumer", migrate)
}
