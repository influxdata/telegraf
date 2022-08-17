//go:build !custom || inputs || inputs.wireless

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/wireless" // register plugin
