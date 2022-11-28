//go:build !custom || outputs || outputs.sensu

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/sensu" // register plugin
