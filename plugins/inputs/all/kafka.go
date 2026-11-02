//go:build !custom || inputs || inputs.kafka

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/kafka" // register plugin
