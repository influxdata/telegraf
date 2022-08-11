//go:build !custom || inputs || inputs.dmcache

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/dmcache" // register plugin
