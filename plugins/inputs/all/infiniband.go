//go:build (!custom || inputs || inputs.infiniband) && linux

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/infiniband" // register plugin
