//go:build !custom || outputs || outputs.aiven_postgresql

package all

import _ "github.com/influxdata/telegraf/plugins/outputs/aiven-postgresql" // register plugin
