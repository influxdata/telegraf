//go:build !custom || (migrations && (inputs || inputs.kube_inventory))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_kube_inventory" // register migration
