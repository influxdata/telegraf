package inputs_openstack

import (
	"fmt"
	"slices"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate the deprecated 'server_diagnotics' option to its
// replacement, adding 'serverdiagnostics' to the 'enabled_services' list. When
// 'enabled_services' is unset the plugin uses a default set of services, so that
// default is materialized before appending to preserve the original behavior.
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Check for the deprecated option and migrate it
	var applied bool
	if rawOldValue, found := plugin["server_diagnotics"]; found {
		applied = true

		// Convert to the actual type
		oldValue, ok := rawOldValue.(bool)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'server_diagnotics'", rawOldValue)
		}

		// Only a 'true' value had an effect, enabling the 'serverdiagnostics'
		// service. A 'false' value is the default and can simply be dropped.
		if oldValue {
			services, _ := plugin["enabled_services"].([]interface{})
			if len(services) == 0 {
				// Mirror the plugin's default set of enabled services
				services = []interface{}{"services", "projects", "hypervisors", "flavors", "networks", "volumes"}
			}
			if !slices.Contains(services, "serverdiagnostics") {
				services = append(services, "serverdiagnostics")
			}
			plugin["enabled_services"] = services
		}

		// Remove the deprecated setting
		delete(plugin, "server_diagnotics")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "openstack")
	cfg.Add("inputs", "openstack", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.openstack", migrate)
}
