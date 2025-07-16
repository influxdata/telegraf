package outputs_mqtt

import (
	"fmt"
	"strings"

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
	var message string
	if rawOldTopicPrefix, found := plugin["topic_prefix"]; found {
		applied = true

		// Convert the options to the actual type
		oldTopicPrefix, ok := rawOldTopicPrefix.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'topic_prefix'", rawOldTopicPrefix)
		}

		// Check if the prefix is already set, otherwise prepend it and inform
		// the users that we modified the topic to make them aware.
		if rawNewTopic, found := plugin["topic"]; found {
			if newTopic, ok := rawNewTopic.(string); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'topic'", rawNewTopic)
			} else if !strings.HasPrefix(newTopic, oldTopicPrefix) && !strings.HasPrefix(newTopic, "/"+oldTopicPrefix) {
				plugin["topic"] = oldTopicPrefix + "/" + newTopic
				message = "added deprecated 'topic_prefix' to existing 'topic'; please check the new setting"
			}
		} else {
			plugin["topic"] = oldTopicPrefix
		}

		// Remove the deprecated option
		delete(plugin, "topic_prefix")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("outputs", "mqtt")
	cfg.Add("outputs", "mqtt", plugin)

	output, err := toml.Marshal(cfg)
	return output, message, err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("outputs.mqtt", migrate)
}
