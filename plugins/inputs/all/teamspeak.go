//go:build !custom || inputs || inputs.teamspeak

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/teamspeak" // register plugin
