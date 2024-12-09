//go:build !custom || inputs || inputs.gnmi

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/gnmi" // register plugin
