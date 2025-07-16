package outputs_kinesis

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

	// Find the partition configuration
	newPartition, err := getPartition(plugin)
	if err != nil {
		return nil, "", err
	}

	// Check for deprecated option(s) and migrate them
	var applied bool
	if rawOldPartitionKey, found := plugin["partitionkey"]; found {
		// Convert the options to the actual type
		oldPartitionKey, ok := rawOldPartitionKey.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'partitionkey'", rawOldPartitionKey)
		}

		// Check if the new setting is present and if so, check if the values are conflicting.
		if rawNewPartitionKey, found := newPartition["key"]; found {
			if newPartitionKey, ok := rawNewPartitionKey.(string); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'partition.key'", rawNewPartitionKey)
			} else if oldPartitionKey != newPartitionKey {
				return nil, "", errors.New("contradicting setting for 'partitionkey' and 'partition.key'")
			}
		}
		applied = true

		// Remove the deprecated option and replace the modified one
		newPartition["key"] = oldPartitionKey
		delete(plugin, "partitionkey")

		// Check for 'use_random_partitionkey' option, which will only be set in a valid way if 'partitionkey' was set
		var oldRandomKey bool
		if rawOldRandomKey, found := plugin["use_random_partitionkey"]; found {
			// Convert the options to the actual type
			oldRandomKey, ok = rawOldRandomKey.(bool)
			if !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'use_random_partitionkey'", rawOldRandomKey)
			}
		} else {
			// If the option is not set, we assume the old value is false
			oldRandomKey = false
		}

		// Check if the new setting is present and if so, check if the values are conflicting.
		if rawNewMethod, found := newPartition["method"]; found {
			if newMethod, ok := rawNewMethod.(string); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'datacenter'", rawNewMethod)
			} else if (oldRandomKey && newMethod != "random") || (!oldRandomKey && newMethod != "static") {
				// If old random key is true, method is expected to be "random"
				// If old random key is false, method is expected to be "static"
				return nil, "", errors.New("contradicting setting for 'use_random_partitionkey' and 'partition.method'")
			}
		}

		// Remove the deprecated option and replace the modified one
		if oldRandomKey {
			newPartition["method"] = "random"
		} else {
			newPartition["method"] = "static"
		}
		delete(plugin, "use_random_partitionkey")
	}

	// Just in case, check if 'use_random_partitionkey' is set without 'partitionkey' somehow
	if _, found := plugin["use_random_partitionkey"]; found {
		delete(plugin, "use_random_partitionkey")
		applied = true
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("outputs", "kinesis")
	cfg.Add("outputs", "kinesis", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

func getPartition(plugin map[string]interface{}) (map[string]interface{}, error) {
	rawPartition := plugin["partition"]
	if rawPartition == nil {
		// Create a new partition if it does not exist
		partition := make(map[string]interface{})
		plugin["partition"] = partition
		return partition, nil
	}

	partition, ok := rawPartition.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type %T for 'partition'", rawPartition)
	}

	return partition, nil
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("outputs.kinesis", migrate)
}
