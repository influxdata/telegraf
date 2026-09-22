//go:build !custom || inputs || inputs.fritzbox_smarthome

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/fritzbox_smarthome" // register plugin
