//go:build !custom || inputs || inputs.cloudwatch

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/cloudwatch" // register plugin
