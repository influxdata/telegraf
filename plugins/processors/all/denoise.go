//go:build !custom || processors || processors.denoise

package all

import _ "github.com/influxdata/telegraf/plugins/processors/denoise" // register plugin
