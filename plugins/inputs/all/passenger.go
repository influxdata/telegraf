//go:build !custom || inputs || inputs.passenger

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/passenger" // register plugin
