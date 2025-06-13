package inputs_sqlserver

import (
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate deprecated SQL Server options
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Check for deprecated options and migrate them
	var applied bool
	var message string

	// Check if database_type is already set - if so, don't override
	_, databaseTypeExists := plugin["database_type"]
	var foundAzureDB bool

	// Migrate azuredb -> database_type
	if azuredbValue, found := plugin["azuredb"]; found {
		applied = true
		foundAzureDB = true

		// Only set database_type if it's not already set (don't overwrite existing)
		if !databaseTypeExists {
			if azuredb, ok := azuredbValue.(bool); ok && azuredb {
				// azuredb = true means Azure SQL Database
				plugin["database_type"] = "AzureSQLDB"
			} else {
				// azuredb = false means on-premises SQL Server
				plugin["database_type"] = "SQLServer"
			}
		}

		// Remove the deprecated setting
		delete(plugin, "azuredb")
	}

	// Migrate query_version -> database_type (only if azuredb wasn't found)
	if _, found := plugin["query_version"]; found {
		applied = true
		if !foundAzureDB {
			// Only set database_type if it's not already set (don't overwrite existing)
			if !databaseTypeExists {
				// For query_version, default to SQLServer regardless of the version
				plugin["database_type"] = "SQLServer"
			}
		}

		// Remove the deprecated setting
		delete(plugin, "query_version")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "sqlserver")
	cfg.Add("inputs", "sqlserver", plugin)
	output, err := toml.Marshal(cfg)
	return output, message, err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.sqlserver", migrate)
}
