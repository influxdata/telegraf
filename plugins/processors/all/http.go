//go:build !custom || processors || processors.http

package all

import _ "github.com/influxdata/telegraf/plugins/processors/http" // register plugin
