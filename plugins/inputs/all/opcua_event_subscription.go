//go:build !custom || inputs || inputs.opcua_event_subscription

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/opcua_event_subscription" // register plugin
