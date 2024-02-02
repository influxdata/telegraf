package inputs_tcp_listener

import (
	"fmt"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

const allowPendingMessagesMsg = `
    Replacement 'inputs.socket_listener' does not allow to configure
    'allowed_pending_messages' and thus the setting is dropped.
`

// Define "old" data structure
type tcpListener map[string]interface{}

// Migration function
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var old tcpListener
	if err := toml.UnmarshalTable(tbl, &old); err != nil {
		return nil, "", err
	}

	// Copy the setting except the special plugin ones to preserve
	// all parser settings of the existing (deprecated) config.
	var msg string
	plugin := make(map[string]interface{}, len(old))
	for k, v := range old {
		switch k {
		case "service_address":
			addr, ok := v.(string)
			if !ok {
				return nil, "", fmt.Errorf("service_address is not a string but %T", v)
			}
			plugin["service_address"] = "tcp://" + addr
		case "allowed_pending_messages":
			msg = allowPendingMessagesMsg
		case "max_tcp_connections":
			plugin["max_connections"] = v
		default:
			plugin[k] = v
		}
	}

	// Create the corresponding metric configurations
	cfg := migrations.CreateTOMLStruct("inputs", "socket_listener")
	cfg.Add("inputs", "socket_listener", plugin)

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
	migrations.AddPluginMigration("inputs.tcp_listener", migrate)
}
