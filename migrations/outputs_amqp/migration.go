package outputs_amqp

import (
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

const messagePrefix = "could not migrate one or more options from the 'outputs.amqp' plugin:"

func migrate(tbl *ast.Table) ([]byte, string, error) {
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	var applied bool
	message := messagePrefix

	if db, found := plugin["database"]; found {
		headers := getHeaders(plugin)

		if _, found := headers["database"]; found {
			message += " 'database' (already set in headers)"
		} else {
			headers["database"] = db.(string)
			delete(plugin, "database")
			applied = true
		}
	}

	if rp, found := plugin["retention_policy"]; found {
		headers := getHeaders(plugin)

		if _, found := headers["retention_policy"]; found {
			message += " 'retention_policy' (already set in headers)"
		} else {
			headers["retention_policy"] = rp.(string)
			delete(plugin, "retention_policy")
			applied = true
		}
	}

	// Delete precision if it exists, as it is no longer used
	if _, found := plugin["precision"]; found {
		applied = true
		delete(plugin, "precision")
	}

	if url, found := plugin["url"]; found {
		brokers := getBrokers(plugin)
		plugin["brokers"] = append(brokers, url.(string))
		delete(plugin, "url")
		applied = true
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("outputs", "amqp")
	cfg.Add("outputs", "amqp", plugin)

	output, err := toml.Marshal(cfg)

	if message == messagePrefix {
		// No options failed to migrate, so we can return an empty message
		return output, "", err
	}
	// Some options failed to migrate, so we return the message
	return output, message, err
}

func getHeaders(plugin map[string]interface{}) map[string]string {
	var headers map[string]string
	if raw, found := plugin["headers"]; found {
		headers = raw.(map[string]string)
	} else {
		headers = make(map[string]string, 1)
		plugin["headers"] = headers
	}
	return headers
}

func getBrokers(plugin map[string]interface{}) []interface{} {
	var brokers []interface{}
	if raw, found := plugin["brokers"]; found {
		brokers = raw.([]interface{})
	} else {
		brokers = make([]interface{}, 1)
		plugin["brokers"] = brokers
	}
	return brokers
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginMigration("outputs.amqp", migrate)
}
