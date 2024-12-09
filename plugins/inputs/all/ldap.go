//go:build !custom || inputs || inputs.ldap

package all

import _ "github.com/influxdata/telegraf/plugins/inputs/ldap" // register plugin
