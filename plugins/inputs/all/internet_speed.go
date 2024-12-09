//go:build !custom || inputs || inputs.internet_speed

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/internet_speed" // register plugin
