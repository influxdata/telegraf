//go:build !custom || inputs || inputs.postgresql

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/postgresql" // register plugin
