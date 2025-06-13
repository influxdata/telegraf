package outputs_wavefront

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

	// Check for deprecated option(s) and migrate them
	var applied bool
	msg := ""

	rawHost, foundHost := plugin["host"]
	rawPort, foundPort := plugin["port"]

	if foundHost {
		host, ok := rawHost.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'host'", rawHost)
		}

		var newURL string
		if foundPort {
			port, ok := rawPort.(int64)
			if !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'port'", rawPort)
			}
			newURL = fmt.Sprintf("http://%s:%d", host, port)
		} else {
			newURL = "http://%s" + host
		}

		if rawURL, found := plugin["url"]; found {
			if url, ok := rawURL.(string); !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'url'", rawURL)
			} else if url != newURL {
				return nil, "", errors.New("contradicting setting for 'host' and 'port' with 'url'")
			}
		}

		applied = true
		delete(plugin, "host")
		delete(plugin, "port")
		plugin["url"] = newURL
	}

	// Cannot automatically migrate 'string_to_number' option, give a message if it is set
	if _, found := plugin["string_to_number"]; found {
		msg = "The 'string_to_number' cannot be migrated automatically and requires manual intervention!"
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, msg, migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("outputs", "wavefront")
	cfg.Add("outputs", "wavefront", plugin)

	output, err := toml.Marshal(cfg)
	return output, msg, err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("outputs.wavefront", migrate)
}
