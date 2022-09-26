//go:build !custom || inputs || inputs.diskio

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/diskio" // register plugin
