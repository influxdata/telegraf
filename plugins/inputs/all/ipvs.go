//go:build !custom || inputs || inputs.ipvs

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/ipvs" // register plugin
