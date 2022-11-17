//go:build !custom || inputs || inputs.apcupsd

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/apcupsd" // register plugin
