package inputs_sflow

import (
	"fmt"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
	"github.com/influxdata/telegraf/migrations/common"
)

const msg = `
    Replacement 'inputs.netflow' will output a different metric format.
	Please adapt your queries!
`

// Define "old" data structure
type sflow struct {
	ServiceAddress string `toml:"service_address"`
	ReadBufferSize string `toml:"read_buffer_size"`
	common.InputOptions
}

// Migration function
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var old sflow
	if err := toml.UnmarshalTable(tbl, &old); err != nil {
		return nil, "", err
	}

	// Fill common options
	plugin := make(map[string]interface{})
	old.InputOptions.Migrate()
	general, err := toml.Marshal(old.InputOptions)
	if err != nil {
		return nil, "", fmt.Errorf("marshalling general options failed: %w", err)
	}
	if err := toml.Unmarshal(general, &plugin); err != nil {
		return nil, "", fmt.Errorf("re-unmarshalling general options failed: %w", err)
	}

	// Use a map for the new plugin and fill in the data
	plugin["service_address"] = old.ServiceAddress
	if old.ReadBufferSize != "" {
		plugin["read_buffer_size"] = old.ReadBufferSize
	}

	// Create the corresponding metric configurations
	cfg := migrations.CreateTOMLStruct("inputs", "netflow")
	cfg.Add("inputs", "netflow", plugin)

	// Marshal the new configuration
	buf, err := toml.Marshal(cfg)
	if err != nil {
		return nil, "", err
	}
	buf = append(buf, []byte("\n")...)

	// Create the new content to output
	return buf, msg, nil
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginMigration("inputs.sflow", migrate)
}
