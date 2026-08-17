//go:build !custom || inputs || inputs.rdap

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/rdap" // register plugin
