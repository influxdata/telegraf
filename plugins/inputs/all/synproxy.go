//go:build !custom || inputs || inputs.synproxy

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/synproxy" // register plugin
