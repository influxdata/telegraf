//go:build !custom || inputs || inputs.infiniband

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/infiniband" // register plugin
