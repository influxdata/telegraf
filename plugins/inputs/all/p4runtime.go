//go:build !custom || inputs || inputs.p4runtime

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/p4runtime" // register plugin
