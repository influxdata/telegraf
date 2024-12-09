//go:build !custom || inputs || inputs.smart

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/smart" // register plugin
