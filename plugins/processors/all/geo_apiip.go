//go:build !custom || processors || processors.geo_apiip

package all

import _ "github.com/influxdata/telegraf/plugins/processors/geo_apiip" // register plugin
