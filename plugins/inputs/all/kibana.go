//go:build !custom || inputs || inputs.kibana

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/kibana" // register plugin
