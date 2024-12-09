//go:build !custom || inputs || inputs.powerdns_recursor

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/powerdns_recursor" // register plugin
