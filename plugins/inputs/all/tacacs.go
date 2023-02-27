//go:build !custom || inputs || inputs.tacacs

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/tacacs" // register plugin
