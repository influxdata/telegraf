//go:build !custom || inputs || inputs.zookeeper

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/zookeeper" // register plugin
