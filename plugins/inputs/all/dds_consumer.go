//go:build !custom || inputs || inputs.dds_consumer

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/dds_consumer" // register plugin
