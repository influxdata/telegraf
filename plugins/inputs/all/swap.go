//go:build !custom || inputs || inputs.swap

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/swap" // register plugin
