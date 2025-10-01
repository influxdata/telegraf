//go:build !custom || processors || processors.xxhash

package all

import _ "github.com/influxdata/telegraf/plugins/processors/xxhash" // register plugin
