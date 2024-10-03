//go:build !custom || inputs || inputs.neoom_beaam

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/neoom_beaam" // register plugin
