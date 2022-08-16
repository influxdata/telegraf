//go:build (!custom || inputs || inputs.varnish) && !windows

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/varnish" // register plugin
