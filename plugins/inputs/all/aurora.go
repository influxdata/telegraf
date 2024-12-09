//go:build !custom || inputs || inputs.aurora

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/aurora" // register plugin
