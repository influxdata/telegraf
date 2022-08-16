//go:build (!custom || inputs || inputs.cgroup) && linux

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/cgroup" // register plugin
