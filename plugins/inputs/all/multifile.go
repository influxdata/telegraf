//go:build !custom || inputs || inputs.multifile

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/multifile" // register plugin
