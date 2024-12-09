//go:build !custom || inputs || inputs.lvm

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/lvm" // register plugin
