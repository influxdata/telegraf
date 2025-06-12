package inputs_docker

import (
	"fmt"

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
	if containerNamesValue, found := plugin["container_names"]; found {
		applied = true

		// Convert to []interface{} for easier handling
		var containerNames []interface{}
		switch v := containerNamesValue.(type) {
		case []interface{}:
			containerNames = v
		case []string:
			for _, name := range v {
				containerNames = append(containerNames, name)
			}
		default:
			return nil, "", fmt.Errorf("container_names value is not a slice: %T", containerNamesValue)
		}

		// Check if container_name_include already exists
		if existingInclude, exists := plugin["container_name_include"]; exists {
			// Merge the arrays
			var existing []interface{}
			switch v := existingInclude.(type) {
			case []interface{}:
				existing = v
			case []string:
				for _, name := range v {
					existing = append(existing, name)
				}
			default:
				return nil, "", fmt.Errorf("container_name_include value is not a slice: %T", existingInclude)
			}

			// Append container_names to existing container_name_include
			merged := append(existing, containerNames...)
			plugin["container_name_include"] = merged
		} else {
			// Create new container_name_include with container_names values
			plugin["container_name_include"] = containerNames
		}

		// Remove deprecated field
		delete(plugin, "container_names")
	}

	// 2. Migrate perdevice -> perdevice_include
	if perdeviceValue, found := plugin["perdevice"]; found {
		// Check if it's a boolean
		perdeviceBool, ok := perdeviceValue.(bool)
		if !ok {
			return nil, "", fmt.Errorf("perdevice value is not a boolean: %T", perdeviceValue)
		}

		// Only apply migration if perdevice=true, since perdevice=false is default behavior
		if perdeviceBool {
			applied = true

			// Check if perdevice_include already exists
			if existingInclude, exists := plugin["perdevice_include"]; exists {
				// perdevice=true means include network and blkio if not already present
				var existing []interface{}
				switch v := existingInclude.(type) {
				case []interface{}:
					existing = v
				case []string:
					for _, name := range v {
						existing = append(existing, name)
					}
				default:
					return nil, "", fmt.Errorf("perdevice_include value is not a slice: %T", existingInclude)
				}

				// Add network and blkio if not present (following plugin's backward compatibility logic)
				hasNetwork := false
				hasBlkio := false
				for _, item := range existing {
					if str, ok := item.(string); ok {
						if str == "network" {
							hasNetwork = true
						}
						if str == "blkio" {
							hasBlkio = true
						}
					}
				}

				if !hasNetwork {
					existing = append(existing, "network")
				}
				if !hasBlkio {
					existing = append(existing, "blkio")
				}

				plugin["perdevice_include"] = existing
			} else {
				// Create new perdevice_include with network and blkio (following plugin logic)
				plugin["perdevice_include"] = []interface{}{"network", "blkio"}
			}
		}

		// Always remove the deprecated field, even if perdevice=false (no migration needed)
		delete(plugin, "perdevice")
		if !perdeviceBool && !applied {
			applied = true
		}
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
