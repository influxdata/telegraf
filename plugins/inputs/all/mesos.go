//go:build !custom || inputs || inputs.mesos

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/mesos" // register plugin
