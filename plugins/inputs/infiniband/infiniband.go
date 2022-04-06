//go:generate go run ../../../tools/generate_plugindata/main.go
//go:generate go run ../../../tools/generate_plugindata/main.go --clean
package infiniband

import (
	"github.com/influxdata/telegraf"
)

// Stores the configuration values for the infiniband plugin - as there are no
// config values, this is intentionally empty
type Infiniband struct {
	Log telegraf.Logger `toml:"-"`
}
