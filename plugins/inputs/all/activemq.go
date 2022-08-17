//go:build !custom || inputs || inputs.activemq

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/activemq" // register plugin
