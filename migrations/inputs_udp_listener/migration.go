package inputs_udp_listener

import (
	"fmt"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	"github.com/influxdata/telegraf/migrations"
)

const allowPendingMessagesMsg = `
    Replacement 'inputs.socket_listener' does not allow to configure
    'allowed_pending_messages' and thus the setting will be dropped.
`

const udpPacketSizeMsg = `
    The deprecated 'udp_buffer_size' setting will be dropped.
`

// Define "old" data structure
type udpListener map[string]any

// Migration function
func migrate(tbl *ast.Table) ([]byte, string, error) {
	// Decode the old data structure
	var old udpListener
	if err := toml.UnmarshalTable(tbl, &old); err != nil {
		return nil, "", err
	}

	// Copy the setting except the special plugin ones to preserve
	// all parser settings of the existing (deprecated) config.
	var msg strings.Builder
	plugin := make(map[string]any, len(old))
	for k, v := range old {
		switch k {
		case "service_address":
			addr, ok := v.(string)
			if !ok {
				return nil, "", fmt.Errorf("service_address is not a string but %T", v)
			}
			plugin["service_address"] = "udp://" + addr
		case "allowed_pending_messages":
			msg.WriteString(allowPendingMessagesMsg)
		case "udp_packet_size":
			msg.WriteString(udpPacketSizeMsg)
		case "udp_buffer_size":
			plugin["read_buffer_size"] = v
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
	return buf, msg.String(), nil
}

// Register the migration function for the plugin type
func init() {
	migrations.AddPluginMigration("inputs.udp_listener", migrate)
}
