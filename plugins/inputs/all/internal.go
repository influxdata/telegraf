//go:build !custom || inputs || inputs.internal

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/internal" // register plugin
