//go:build !custom || inputs || inputs.exec

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/exec" // register plugin
