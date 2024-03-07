//go:build !custom || processors || processors.apiip

package all

import _ "github.com/influxdata/telegraf/plugins/processors/geo_apiip" // register plugin
