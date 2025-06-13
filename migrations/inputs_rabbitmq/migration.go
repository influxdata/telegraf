package inputs_rabbitmq

import (
	"errors"
	"fmt"
	"slices"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate deprecated RabbitMQ options
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Check for deprecated options and migrate them
	var applied bool
	var message string

	// Migrate the deprecated "name" option to tags
	if rawOldName, found := plugin["name"]; found {
		applied = true

		oldName, ok := rawOldName.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'name'", rawOldName)
		}

		// Merge with potentially existing tags
		var tags map[string]interface{}
		if rawTags, found := plugin["tags"]; found {
			if tags, ok = rawTags.(map[string]interface{}); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'tags'", rawTags)
			} else if rawName, found := tags["name"]; found {
				if name, ok := rawName.(string); !ok {
					return nil, "", fmt.Errorf("unexpected type %T for 'name' tag", rawName)
				} else if name != oldName {
					return nil, "", errors.New("contradicting setting for 'name' and 'name' tag")
				}
			} else {
				tags["name"] = oldName
			}
		} else {
			tags = map[string]interface{}{"name": oldName}
		}

		// Remove the deprecated setting
		plugin["tags"] = tags
		delete(plugin, "name")
	}

	// Migrate queues -> queue_name_include
	if rawOldQueues, found := plugin["queues"]; found {
		applied = true

		oldQueues, err := migrations.AsStringSlice(rawOldQueues)
		if err != nil {
			return nil, "", fmt.Errorf("setting 'queues': %w", err)
		}

		// Merge with potentially existing setting
		var includes []string
		if rawNewQueueNameInclude, found := plugin["queue_name_include"]; found {
			if includes, err = migrations.AsStringSlice(rawNewQueueNameInclude); err != nil {
				return nil, "", fmt.Errorf("setting 'queue_name_include': %w", err)
			}
		}

		for _, q := range oldQueues {
			if !slices.Contains(includes, q) {
				includes = append(includes, q)
			}
		}

		// Remove the deprecated setting
		plugin["queue_name_include"] = includes
		delete(plugin, "queues")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "rabbitmq")
	cfg.Add("inputs", "rabbitmq", plugin)

	output, err := toml.Marshal(cfg)
	return output, message, err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.rabbitmq", migrate)
}
