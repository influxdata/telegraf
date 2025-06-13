package inputs_statsd

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

	// Remove option as it's being ignored
	if _, found := plugin["udp_packet_size"]; found {
		applied = true
		delete(plugin, "udp_packet_size")
	}

	if rawOldParseDataDogTags, found := plugin["parse_data_dog_tags"]; found {
		applied = true

		// Convert the options to the actual type
		oldParseDataDogTags, ok := rawOldParseDataDogTags.(bool)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'parse_data_dog_tags'", rawOldParseDataDogTags)
		}

		// Check if the new setting is present and if so, check if the values are
		// conflicting.
		if rawNewDataDogExtensions, found := plugin["datadog_extensions"]; found {
			if newDataDogExtensions, ok := rawNewDataDogExtensions.(bool); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'datadog_extensions'", rawNewDataDogExtensions)
			} else if oldParseDataDogTags != newDataDogExtensions {
				return nil, "", errors.New("contradicting setting for 'parse_data_dog_tags' and 'datadog_extensions'")
			}
		} else if oldParseDataDogTags {
			// Only set the option if it is not the default for the new option
			plugin["datadog_extensions"] = true
		}

		// Remove the deprecated option and replace the modified one
		delete(plugin, "parse_data_dog_tags")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "statsd")
	cfg.Add("inputs", "statsd", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.statsd", migrate)
}
