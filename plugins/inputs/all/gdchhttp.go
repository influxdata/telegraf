//go:build !custom || inputs || inputs.gdchhttp

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/gdchhttp" // register plugin
