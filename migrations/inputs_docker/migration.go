package inputs_docker

import (
	"errors"
	"fmt"
	"slices"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate multiple deprecated Docker options
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	var applied bool

	// 1. Migrate container_names -> container_name_include
	if rawContainerNames, found := plugin["container_names"]; found {
		applied = true

		// Convert to actual type
		containerNames, err := migrations.AsStringSlice(rawContainerNames)
		if err != nil {
			return nil, "", fmt.Errorf("setting 'container_names': %w", err)
		}

		// Check if container_name_include already exists
		var includes []string
		if rawContainerNameInclude, found := plugin["container_name_include"]; found {
			// Convert to actual type
			if includes, err = migrations.AsStringSlice(rawContainerNameInclude); err != nil {
				return nil, "", fmt.Errorf("setting 'container_name_include': %w", err)
			}
		}

		// Merge the options
		for _, name := range containerNames {
			if !slices.Contains(includes, name) {
				includes = append(includes, name)
			}
		}

		// Remove deprecated field and replace by the migrated one
		plugin["container_name_include"] = includes
		delete(plugin, "container_names")
	}

	// 2. Migrate perdevice -> perdevice_include
	if rawPerDevice, found := plugin["perdevice"]; found {
		applied = true

		// Check if it's a boolean
		perDevice, ok := rawPerDevice.(bool)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'perdevice'", rawPerDevice)
		}

		// Get the existing include list for checking if set
		var includes []string
		if rawPerDeviceInclude, found := plugin["perdevice_include"]; found {
			var err error
			if includes, err = migrations.AsStringSlice(rawPerDeviceInclude); err != nil {
				return nil, "", fmt.Errorf("setting 'perdevice_include': %w", err)
			}
			if perDevice {
				if !slices.Contains(includes, "network") {
					includes = append(includes, "network")
				}
				if !slices.Contains(includes, "blkio") {
					includes = append(includes, "blkio")
				}
			} else if slices.Contains(includes, "network") || slices.Contains(includes, "blkio") {
				return nil, "", errors.New("contradicting settings for 'perdevice' and 'perdevice_include'")
			}
		} else if perDevice {
			includes = []string{"cpu", "network", "blkio"}
		}

		// Remove deprecated setting and add new one
		plugin["perdevice_include"] = includes
		delete(plugin, "perdevice")
	}

	// 3. Migrate total -> total_include
	if totalValue, found := plugin["total"]; found {
		// Check if it's a boolean
		totalBool, ok := totalValue.(bool)
		if !ok {
			return nil, "", fmt.Errorf("total value is not a boolean: %T", totalValue)
		}

		// Always remove the deprecated field
		delete(plugin, "total")
		applied = true

		// Only modify total_include if total=false AND there's an existing total_include that needs modification
		if !totalBool {
			if existingInclude, exists := plugin["total_include"]; exists {
				// total=false means only cpu (following plugin logic)
				var existing []interface{}
				switch v := existingInclude.(type) {
				case []interface{}:
					existing = v
				case []string:
					for _, name := range v {
						existing = append(existing, name)
					}
				default:
					return nil, "", fmt.Errorf("total_include value is not a slice: %T", existingInclude)
				}

				// Keep only cpu if present, remove others
				var cpuOnly []interface{}
				for _, item := range existing {
					if str, ok := item.(string); ok && str == "cpu" {
						cpuOnly = append(cpuOnly, "cpu")
						break
					}
				}

				plugin["total_include"] = cpuOnly
			}
			// Don't create total_include for total=false - let plugin handle defaults
		}
		// total=true is default behavior - no action needed
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "docker")
	cfg.Add("inputs", "docker", plugin)

	output, err := toml.Marshal(cfg)
	return output, "", err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.docker", migrate)
}
