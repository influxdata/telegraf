//go:build !custom || processors || processors.clone

package all

import _ "github.com/influxdata/telegraf/plugins/processors/clone" // register plugin
