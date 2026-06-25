//go:build !custom || (migrations && (inputs || inputs.openstack))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_openstack" // register migration
