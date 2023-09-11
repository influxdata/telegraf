//go:build !custom || inputs || inputs.lustre2_lctl

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/lustre2_lctl" // register plugin
