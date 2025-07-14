//go:build !custom || (migrations && (inputs || inputs.openldap))

package all

import _ "github.com/influxdata/telegraf/migrations/inputs_openldap" // register migration
