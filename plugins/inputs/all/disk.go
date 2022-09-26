//go:build !custom || inputs || inputs.disk

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/disk" // register plugin
