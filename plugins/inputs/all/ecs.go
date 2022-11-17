//go:build !custom || inputs || inputs.ecs

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/ecs" // register plugin
