//go:build !custom || inputs || inputs.hugepages

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/hugepages" // register plugin
