//go:build (!custom || inputs || inputs.ras) && linux && (386 || amd64 || arm || arm64)

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/ras" // register plugin
