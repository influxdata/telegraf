package outputs_influxdb

import (
	"errors"
	"fmt"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/internal/choice"
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
	if oldURL, found := plugin["url"]; found {
		applied = true

		var urls []string
		// Merge the old URL and the new URLs with deduplication
		if newURLs, found := plugin["urls"]; found {
			list, ok := newURLs.([]interface{})
			if !ok {
				return nil, "", errors.New("'urls' setting is not a list")
			}
			for _, raw := range list {
				nu, ok := raw.(string)
				if !ok {
					return nil, "", fmt.Errorf("unexpected 'urls' entry %v (%T)", raw, raw)
				}
				urls = append(urls, nu)
			}
		}
		ou, ok := oldURL.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected 'url' entry %v (%T)", ou, ou)
		}

		if !choice.Contains(ou, urls) {
			urls = append(urls, ou)
		}

		// Update replacement and remove the deprecated setting
		plugin["urls"] = urls
		delete(plugin, "url")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("outputs", "influxdb")
	cfg.Add("outputs", "influxdb", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("outputs.influxdb", migrate)
}
