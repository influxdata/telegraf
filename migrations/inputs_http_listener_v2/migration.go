package inputs_http_listener_v2

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strconv"

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

	if rawOldPath, found := plugin["path"]; found {
		applied = true

		// Convert the options to the actual type
		oldPath, ok := rawOldPath.(string)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'path'", rawOldPath)
		}

		// Merge the option with the replacement
		var paths []string
		if rawNewPaths, found := plugin["paths"]; found {
			var err error
			paths, err = migrations.AsStringSlice(rawNewPaths)
			if err != nil {
				return nil, "", fmt.Errorf("'paths' option: %w", err)
			}
		}

		if !slices.Contains(paths, oldPath) {
			paths = append(paths, oldPath)
		}

		// Remove the deprecated option and replace the modified one
		delete(plugin, "path")
		plugin["paths"] = paths
	}

	if rawOldPort, found := plugin["port"]; found {
		applied = true

		// Convert the options to the actual type
		oldPort, ok := rawOldPort.(int64)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'port'", rawOldPort)
		}

		// Check if the new setting is present and if so, check if the values are
		// conflicting.
		var address string
		if rawNewServiceAddress, found := plugin["service_address"]; found {
			newServiceAddress, ok := rawNewServiceAddress.(string)
			if !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'service_address'", rawNewServiceAddress)
			}
			address = newServiceAddress
		}

		// Check if the port of the service-address matches
		u, err := url.Parse(address)
		if err != nil {
			return nil, "", fmt.Errorf("parsing 'service_address' failed: %w", err)
		}

		if u.Scheme == "" {
			u.Scheme = "tcp"
		}

		if u.Port() != "" {
			newPort, err := strconv.ParseInt(u.Port(), 10, 64)
			if err != nil {
				return nil, "", fmt.Errorf("converting port of 'service_address' failed: %w", err)
			}
			if newPort != oldPort {
				return nil, "", errors.New("contradicting setting for 'port' and port of 'service_address'")
			}
		} else {
			u.Host = fmt.Sprintf("%s:%d", u.Host, oldPort)
			address = u.String()
		}

		// Remove the deprecated option and replace the modified one
		delete(plugin, "port")
		plugin["service_address"] = address
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configurations
	cfg := migrations.CreateTOMLStruct("inputs", "http_listener_v2")
	cfg.Add("inputs", "http_listener_v2", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.http_listener_v2", migrate)
}
