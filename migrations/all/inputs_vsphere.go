//go:build !custom || (migrations && (inputs || inputs.vsphere))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_vsphere" // register migration
