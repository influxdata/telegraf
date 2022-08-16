//go:build (!custom || inputs || inputs.slab) && linux

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/slab" // register plugin
