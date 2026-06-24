package inputs_tail

import (
	"fmt"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

// Migration function to migrate the deprecated 'from_beginning' option to its
// replacement 'initial_read_offset'.
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var plugin map[string]interface{}
	if err := toml.UnmarshalTable(tbl, &plugin); err != nil {
		return nil, "", err
	}

	// Check for the deprecated option and migrate it
	var applied bool
	var message string
	if rawOldValue, found := plugin["from_beginning"]; found {
		applied = true

		// Convert to the actual type
		oldValue, ok := rawOldValue.(bool)
		if !ok {
			return nil, "", fmt.Errorf("unexpected type %T for 'from_beginning'", rawOldValue)
		}

		// A 'from_beginning' value of 'true' corresponds to reading from the
		// 'beginning' while 'false' corresponds to the default 'saved-or-end'
		// behavior.
		expected := "saved-or-end"
		if oldValue {
			expected = "beginning"
		}

		// The 'initial_read_offset' option supersedes 'from_beginning' and takes
		// precedence at runtime. If it is already set, keep it and just drop the
		// deprecated option, but warn the user if the two settings disagree.
		if rawNewValue, found := plugin["initial_read_offset"]; found {
			newValue, ok := rawNewValue.(string)
			if !ok {
				return nil, "", fmt.Errorf("unexpected type %T for 'initial_read_offset'", rawNewValue)
			}
			if newValue != expected {
				message = fmt.Sprintf("ignoring 'from_beginning = %v' as it conflicts with 'initial_read_offset = %q'", oldValue, newValue)
			}
		} else {
			plugin["initial_read_offset"] = expected
		}

		// Remove the deprecated setting
		delete(plugin, "from_beginning")
	}

	// No options migrated so we can exit early
	if !applied {
		return nil, "", migrations.ErrNotApplicable
	}

	// Create the corresponding plugin configuration
	cfg := migrations.CreateTOMLStruct("inputs", "tail")
	cfg.Add("inputs", "tail", plugin)

	output, err := toml.Marshal(cfg)
	return output, message, err
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginOptionMigration("inputs.tail", migrate)
}
