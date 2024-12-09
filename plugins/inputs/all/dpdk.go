//go:build !custom || inputs || inputs.dpdk

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/dpdk" // register plugin
