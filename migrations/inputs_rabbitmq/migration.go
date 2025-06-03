package inputs_rabbitmq

import (
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

	// Remove a deprecated name option (it should be migrated to global tags, not plugin tags)
	// The name option is deprecated in favor of using Telegraf's global tags configuration
	if _, found := plugin["name"]; found {
		applied = true
		// Remove the deprecated setting - users should use global tags instead
		delete(plugin, "name")
		message = "removed deprecated 'name' option; use global 'tags' configuration instead"
	}

	// Migrate queues -> queue_name_include
	if queuesValue, found := plugin["queues"]; found {
		applied = true

		// Only set queue_name_include if it's not already set (don't overwrite existing)
		if _, queueIncludeExists := plugin["queue_name_include"]; !queueIncludeExists {
			plugin["queue_name_include"] = queuesValue
		}

		// Remove the deprecated setting
		delete(plugin, "queues")

		if message != "" {
			message += "; migrated 'queues' option to 'queue_name_include'"
		} else {
			message = "migrated 'queues' option to 'queue_name_include'"
		}
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
