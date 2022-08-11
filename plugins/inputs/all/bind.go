//go:build !custom || inputs || inputs.bind

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/bind" // register plugin
