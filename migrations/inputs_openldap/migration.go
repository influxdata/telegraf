package inputs_openldap

import (
	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate deprecated SSL options to TLS options
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Check for deprecated options and migrate them
	var applied bool
	var message string

	// Migrate ssl -> tls
	if sslValue, found := plugin["ssl"]; found {
		applied = true
		// Only set tls if it's not already set (don't overwrite existing tls setting)
		if _, tlsExists := plugin["tls"]; !tlsExists {
			plugin["tls"] = sslValue
		}
		// Remove the deprecated setting
		delete(plugin, "ssl")
		message = "migrated 'ssl' option to 'tls'"
	}

	// Migrate ssl_ca -> tls_ca
	if sslCAValue, found := plugin["ssl_ca"]; found {
		applied = true
		// Only set tls_ca if it's not already set (don't overwrite existing tls_ca setting)
		if _, tlsCAExists := plugin["tls_ca"]; !tlsCAExists {
			plugin["tls_ca"] = sslCAValue
		}
		// Remove the deprecated setting
		delete(plugin, "ssl_ca")
		if message != "" {
			message += "; migrated 'ssl_ca' option to 'tls_ca'"
		} else {
			message = "migrated 'ssl_ca' option to 'tls_ca'"
		}
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "openldap")
	cfg.Add("inputs", "openldap", plugin)

	output, err := toml.Marshal(cfg)
	return output, message, err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.openldap", migrate)
}
