//go:build !custom || inputs || inputs.xtremio

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/xtremio" // register plugin
