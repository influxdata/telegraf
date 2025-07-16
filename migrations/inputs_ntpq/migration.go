package inputs_ntpq

import (
	"fmt"
	"slices"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"github.com/kballard/go-shellquote"

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
	if rawOldDNSLookup, found := plugin["dns_lookup"]; found {
		applied = true

		// Convert the options to the actual type
		oldDNSLookup, ok := rawOldDNSLookup.(bool)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'dns_lookup'", rawOldDNSLookup)
		}

		// Only insert options if dns_lookup was actually set to false
		if !oldDNSLookup {
			// Merge the option with the replacement
			var options []string
			if rawOptions, found := plugin["options"]; found {
				opts, ok := rawOptions.(string)
				if !ok {
					return nil, "", fmt.Errorf("unexpected type %T for 'options'", rawOptions)
				}
				o, err := shellquote.Split(opts)
				if err != nil {
					return nil, "", fmt.Errorf("splitting 'options' failed: %w", err)
				}
				options = o
			} else {
				options = append(options, "-p")
			}

			if !slices.Contains(options, "-n") {
				options = append(options, "-n")
			}
			plugin["options"] = strings.Join(options, " ")
		}

		// Remove the deprecated option
		delete(plugin, "dns_lookup")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "ntpq")
	cfg.Add("inputs", "ntpq", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.ntpq", migrate)
}
