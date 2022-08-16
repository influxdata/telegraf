//go:build (!custom || inputs || inputs.zfs) && (linux || freebsd)

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/zfs" // register plugin
