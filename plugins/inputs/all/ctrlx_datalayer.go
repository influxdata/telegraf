//go:build !custom || inputs || inputs.ctrlx_datalayer

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/ctrlx_datalayer" // register plugin
